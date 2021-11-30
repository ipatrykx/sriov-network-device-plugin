package utils

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
)

const (
	// fwAppNameKey to extract DDP name
	fwAppNameKey = "fw.app.name"

	pciBus = "pci"
)

// IsDevlinkSupportedByPCIDevice checks if PCI devie supports devlink info command
func IsDevlinkSupportedByPCIDevice(device string) bool {
	return IsDevlinkSupportedByDevice(pciBus, device)
}

// IsDevlinkSupportedByDevice checks if device supports devlink info command
func IsDevlinkSupportedByDevice(bus, device string) bool {
	if _, err := DevlinkGetDeviceInfoByNameAndKey(bus, device, fwAppNameKey); err != nil {
		return false
	}

	return true
}

// DevlinkGetDDPProfiles returns DDP for selected device
func DevlinkGetDDPProfiles(device string) (string, error) {
	return DevlinkGetDeviceInfoByNameAndKey(pciBus, device, fwAppNameKey)
}

// DevlinkGetDeviceInfoByNameAndKeys returns values for selected keys in device info
func DevlinkGetDeviceInfoByNameAndKeys(bus, device string,
	keys []string) (map[string]string, error) {
	data, err := netlink.DevlinkGetDeviceInfoByNameAsMap(bus, device)
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
func DevlinkGetDeviceInfoByNameAndKey(bus, device, key string) (string, error) {
	keys := []string{key}
	info, err := DevlinkGetDeviceInfoByNameAndKeys(bus, device, keys)
	if err != nil {
		return "", err
	}

	if value, exists := info[key]; exists {
		return value, nil
	}

	return "", KeyNotFoundError("DevlinkGetDeviceInfoByNameAndKey", key)
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
