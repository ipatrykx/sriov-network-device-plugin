package utils

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
	"golang.org/x/sys/unix"
)

const (
	// DevlinkCmdInfoGet command ID
	DevlinkCmdInfoGet = 51 // DEVLINK_CMD_INFO_GET
	// DevlinkAttrInfoVersionFixed is nested structure for fixed values
	DevlinkAttrInfoVersionFixed = 100 // DEVLINK_ATTR_INFO_VERSION_FIXED
	// DevlinkAttrInfoVersionRunning is nested structure for current values
	DevlinkAttrInfoVersionRunning = 101 // DEVLINK_ATTR_INFO_VERSION_RUNNING
	// DevlinkAttrInfoVersionStored is nested structure for stored values
	DevlinkAttrInfoVersionStored = 102 // DEVLINK_ATTR_INFO_VERSION_STORED
	// DevlinkAttrInfoVersionName is value that indicates info label
	DevlinkAttrInfoVersionName = 103 // DEVLINK_ATTR_INFO_VERSION_NAME
	// DevlinkAttrInfoVersionValue is value that indicates info value
	DevlinkAttrInfoVersionValue = 104 // DEVLINK_ATTR_INFO_VERSION_VALUE
	// FwAppNameKey to extract DDP name
	FwAppNameKey = "fw.app.name"

	pciBus        = "pci"
	headerSize    = 4
	nestedAttrNum = 2
)

// devlinkInfoGetter is function that is responsible for getting devlink info message
type devlinkInfoGetter func(bus, device string) ([]byte, error)

// DevlinkInfoClient is a Devlink Info Client
type DevlinkInfoClient struct {
	infoGetter devlinkInfoGetter
}

// NewDevlinkInfoClient - returns new client
func NewDevlinkInfoClient() DevlinkInfoClient {
	return newDevlinkInfoClient(getDevlinkInfo)
}

func newDevlinkInfoClient(infoGetter devlinkInfoGetter) DevlinkInfoClient {
	var dic DevlinkInfoClient
	dic.infoGetter = infoGetter

	return dic
}

// IsDevlinkSupportedByPCIDevice checks if PCI devie supports devlink info command
func (dic DevlinkInfoClient) IsDevlinkSupportedByPCIDevice(device string) bool {
	return dic.IsDevlinkSupportedByDevice(pciBus, device)
}

// IsDevlinkSupportedByDevice checks if device supports devlink info command
func (dic DevlinkInfoClient) IsDevlinkSupportedByDevice(bus, device string) bool {
	if _, err := dic.DevlinkGetDeviceInfoByNameAndKey(bus, device, FwAppNameKey); err != nil {
		return false
	}

	return true
}

// DevlinkGetDDPProfiles returns DDP for selected device
func (dic DevlinkInfoClient) DevlinkGetDDPProfiles(device string) (string, error) {
	return dic.DevlinkGetDeviceInfoByNameAndKey(pciBus, device, FwAppNameKey)
}

// DevlinkGetDeviceInfoByNameAndKeys returns values for selected keys in device info
func (dic DevlinkInfoClient) DevlinkGetDeviceInfoByNameAndKeys(bus, device string,
	keys []string) (map[string]string, error) {
	data, err := dic.DevlinkGetDeviceInfoByName(bus, device)
	if err != nil {
		return nil, err
	}

	info := make(map[string]string)

	for _, key := range keys {
		if value, exists := data[key]; exists {
			info[key] = value
		} else {
			return nil, KeyNotFoundError("DevlinkGetDeviceInfoByNameAndKeys", key)
		}
	}

	return info, nil
}

// DevlinkGetDeviceInfoByNameAndKey returns values for selected key in device info
func (dic DevlinkInfoClient) DevlinkGetDeviceInfoByNameAndKey(bus, device, key string) (string, error) {
	keys := []string{key}
	info, err := dic.DevlinkGetDeviceInfoByNameAndKeys(bus, device, keys)
	if err != nil {
		return "", err
	}

	if value, exists := info[key]; exists {
		return value, nil
	}

	return "", KeyNotFoundError("DevlinkGetDeviceInfoByNameAndKey", key)
}

// DevlinkGetDeviceInfoByName returns devlink info for selected device
func (dic DevlinkInfoClient) DevlinkGetDeviceInfoByName(bus, device string) (map[string]string, error) {
	response, err := dic.infoGetter(bus, device)
	if err != nil {
		return nil, err
	}

	info, err := parseInfoMsg(response)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func getDevlinkInfo(bus, device string) ([]byte, error) {
	req, err := createCmdReq(DevlinkCmdInfoGet, bus, device)
	if err != nil {
		return nil, err
	}

	response, err := req.Execute(unix.NETLINK_GENERIC, 0)
	if err != nil {
		return nil, err
	}

	if len(response) < 1 {
		return nil, ErrMessageTooShort
	}

	return response[0], nil
}

func parseInfoMsg(msg []byte) (map[string]string, error) {
	if len(msg) < headerSize {
		return nil, ErrMessageTooShort
	}

	info := make(map[string]string)
	err := parseInfoData(msg[headerSize:], info)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func getNestedInfoData(msg []byte) (string, string, error) {
	nestedAttrs, err := nl.ParseRouteAttr(msg)

	if err != nil {
		return "", "", err
	}

	if len(nestedAttrs) != nestedAttrNum {
		return "", "", ReadAttributesError("getNestedInfoData", "too few attributes in nested structure")
	}

	var key, value string

	for _, nestedAttr := range nestedAttrs {
		switch nestedAttr.Attr.Type {
		case DevlinkAttrInfoVersionName:
			key = strings.ReplaceAll(string(nestedAttr.Value), "\x00", "")
			key = strings.TrimSpace(key)
		case DevlinkAttrInfoVersionValue:
			value = strings.ReplaceAll(string(nestedAttr.Value), "\x00", "")
			value = strings.TrimSpace(value)
		}
	}

	if key == "" {
		return "", "", ReadAttributesError("getNestedInfoData", "key not found")
	}

	if value == "" {
		return "", "", ReadAttributesError("getNestedInfoData", "value not found")
	}

	return key, value, nil
}

func parseInfoData(msg []byte, data map[string]string) error {
	attrs, err := nl.ParseRouteAttr(msg)
	if err != nil {
		return err
	}

	for _, attr := range attrs {
		switch attr.Attr.Type {
		case DevlinkAttrInfoVersionRunning, DevlinkAttrInfoVersionFixed, DevlinkAttrInfoVersionStored:
			key, value, err := getNestedInfoData(attr.Value)
			if err != nil {
				return err
			}
			data[key] = value
		}
	}

	if len(data) == 0 {
		return ReadAttributesError("parseInfoData", "no data found")
	}

	return nil
}

func createCmdReq(cmd uint8, bus, device string) (*nl.NetlinkRequest, error) {
	f, err := netlink.GenlFamilyGet(nl.GENL_DEVLINK_NAME)
	if err != nil {
		return nil, err
	}

	req := nl.NewNetlinkRequest(int(f.ID), unix.NLM_F_REQUEST|unix.NLM_F_ACK)

	msg := &nl.Genlmsg{
		Command: cmd,
		Version: nl.GENL_DEVLINK_VERSION,
	}
	req.AddData(msg)

	b := make([]byte, len(bus)+1)
	copy(b, bus)
	data := nl.NewRtAttr(nl.DEVLINK_ATTR_BUS_NAME, b)
	req.AddData(data)

	b = make([]byte, len(device)+1)
	copy(b, device)
	data = nl.NewRtAttr(nl.DEVLINK_ATTR_DEV_NAME, b)
	req.AddData(data)

	return req, nil
}

// ErrKeyNotFound error when key is not found in the parsed response
var ErrKeyNotFound = errors.New("key could not be found")

// ErrMessageTooShort error when netlink message is too short
var ErrMessageTooShort = errors.New("message too short")

// ErrReadAttributes error when netlink message attributes cannot be read
var ErrReadAttributes = errors.New("could not read attributes")

// ErrDDPNotSupported error when device does not support DDP
var ErrDDPNotSupported = errors.New("this device seems not to support DDP")

// KeyNotFoundError returns ErrKeyNotFound
func KeyNotFoundError(function, key string) error {
	return fmt.Errorf("%s - %w: %s", function, ErrKeyNotFound, key)
}

// ReadAttributesError returns ErrReadAttributes
func ReadAttributesError(function, message string) error {
	return fmt.Errorf("%s - %w: %s", function, ErrReadAttributes, message)
}

// DDPNotSupportedError returns ErrDDPNotSupported
func DDPNotSupportedError(device string) error {
	return fmt.Errorf("%w: %s", ErrDDPNotSupported, device)
}
