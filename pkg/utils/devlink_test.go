package utils

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("In the utils package", func() {
	DescribeTable("Parse devlink response",
		func(shouldNotFail bool, response []byte, expectedInfo map[string]string) {
			m, err := parseInfoMsg(response)
			if shouldNotFail {
				Expect(err).NotTo(HaveOccurred())
				for key, value := range m {
					Expect(expectedInfo[key]).Should(Equal(value))
				}
			} else {
				Expect(err).To(HaveOccurred())
			}

		},
		Entry("Valid devlink response",
			true,
			devlinkInfo(),
			devlinkTestInfoParesd(),
		),
		Entry("Invalid devlink response",
			false,
			[]byte{},
			map[string]string{},
		),
	)

	Describe("creating netlink request", func() {
		Context("create netlink request with dummy data", func() {
			It("should create proper request", func() {
				_, err := createCmdReq(DevlinkCmdInfoGet, "pci", "0000:00:00.0")
				Expect(err).NotTo(HaveOccurred())
			},
			)
		},
		)
	},
	)

	Describe("Test DDP - devlink", func() {
		Context("testing devlink requests with proper devlink response from NIC", func() {
			var client DevlinkInfoClient
			var info map[string]string
			BeforeEach(func() {
				client = newDevlinkInfoClient(mockDevlinkInfoGetter)
				info = devlinkTestInfoParesd()
			})
			It("tests if fw.app.name can be obtained", func() {
				name, err := client.DevlinkGetDDPProfiles("")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(name).To(Equal(info["fw.app.name"]))
			})
			It("tests if devlink support can be determined", func() {
				isSupported := client.IsDevlinkSupportedByPCIDevice("0000:00:00.0")
				Expect(isSupported).Should(BeTrue())
			})
			It("tests if multiple key-value pairs can be obtained from response", func() {
				keys := []string{"fw.app", "fw.mgmt.api", "fw.undi"}
				values, err := client.DevlinkGetDeviceInfoByNameAndKeys("", "", keys)
				Expect(err).ShouldNot(HaveOccurred())
				for k, v := range values {
					Expect(v).To(Equal(info[k]))
				}
			})
			It("tests if single key-value pair can be obtained from the response", func() {
				key := "fw.bundle_id"
				value, err := client.DevlinkGetDeviceInfoByNameAndKey("", "", key)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(value).To(Equal(info[key]))
			})
			It("tests if all key-value pairs are obtained correctly", func() {
				values, err := client.DevlinkGetDeviceInfoByName("", "")
				Expect(err).ShouldNot(HaveOccurred())
				for k, v := range values {
					Expect(v).To(Equal(info[k]))
				}
			})
		})

		Context("testing devlink requests with empty (erroneous) devlink response from NIC", func() {
			var client DevlinkInfoClient
			var info map[string]string
			BeforeEach(func() {
				client = newDevlinkInfoClient(mockDevlinkInfoGetterEmpty)
				info = devlinkTestInfoParesd()
			})
			It("tests error handling while getting DDP profile name", func() {
				name, err := client.DevlinkGetDDPProfiles("")
				Expect(err).Should(HaveOccurred())
				Expect(name).NotTo(Equal(info["fw.app.name"]))
			})
			It("tests error handling when checking DDP support", func() {
				isSupported := client.IsDevlinkSupportedByPCIDevice("0000:00:00.0")
				Expect(isSupported).Should(BeFalse())
			})
			It("tests error handling when obtaining multiple key-value pairs", func() {
				keys := []string{"fw.app", "fw.mgmt.api", "fw.undi"}
				values, err := client.DevlinkGetDeviceInfoByNameAndKeys("", "", keys)
				Expect(err).Should(HaveOccurred())
				for k, v := range values {
					Expect(v).NotTo(Equal(info[k]))
				}
			})
			It("tests error handling when obtaining single ky-value pair", func() {
				key := "fw.bundle_id"
				value, err := client.DevlinkGetDeviceInfoByNameAndKey("", "", key)
				Expect(err).Should(HaveOccurred())
				Expect(value).NotTo(Equal(info[key]))
			})
			It("tests error handling when obtaining all key-value pairs", func() {
				values, err := client.DevlinkGetDeviceInfoByName("", "")
				Expect(err).Should(HaveOccurred())
				for k, v := range values {
					Expect(v).NotTo(Equal(info[k]))
				}
			})
		})
	},
	)
})

func mockDevlinkInfoGetter(bus, device string) ([]byte, error) {
	return devlinkInfo(), nil
}

func mockDevlinkInfoGetterEmpty(bus, device string) ([]byte, error) {
	return []byte{}, nil
}

func devlinkInfo() []byte {
	// This data was obtained from Intel E810-C NIC
	return []byte{51, 1, 0, 0, 8, 0, 1, 0, 112, 99, 105, 0, 17, 0, 2, 0, 48, 48,
		48, 48, 58, 56, 52, 58, 48, 48, 46, 48, 0, 0, 0, 0, 8, 0, 98, 0, 105, 99,
		101, 0, 28, 0, 99, 0, 51, 48, 45, 56, 57, 45, 97, 51, 45, 102, 102, 45,
		102, 102, 45, 99, 97, 45, 48, 53, 45, 54, 56, 0, 36, 0, 100, 0, 13, 0, 103,
		0, 98, 111, 97, 114, 100, 46, 105, 100, 0, 0, 0, 0, 15, 0, 104, 0, 75, 56,
		53, 53, 56, 53, 45, 48, 48, 48, 0, 0, 28, 0, 101, 0, 12, 0, 103, 0, 102, 119,
		46, 109, 103, 109, 116, 0, 10, 0, 104, 0, 53, 46, 52, 46, 53, 0, 0, 0, 28, 0,
		101, 0, 16, 0, 103, 0, 102, 119, 46, 109, 103, 109, 116, 46, 97, 112, 105, 0,
		8, 0, 104, 0, 49, 46, 55, 0, 40, 0, 101, 0, 18, 0, 103, 0, 102, 119, 46, 109,
		103, 109, 116, 46, 98, 117, 105, 108, 100, 0, 0, 0, 15, 0, 104, 0, 48, 120,
		51, 57, 49, 102, 55, 54, 52, 48, 0, 0, 32, 0, 101, 0, 12, 0, 103, 0, 102, 119,
		46, 117, 110, 100, 105, 0, 13, 0, 104, 0, 49, 46, 50, 56, 57, 56, 46, 48, 0,
		0, 0, 0, 32, 0, 101, 0, 16, 0, 103, 0, 102, 119, 46, 112, 115, 105, 100, 46,
		97, 112, 105, 0, 9, 0, 104, 0, 50, 46, 52, 50, 0, 0, 0, 0, 40, 0, 101, 0, 17,
		0, 103, 0, 102, 119, 46, 98, 117, 110, 100, 108, 101, 95, 105, 100, 0, 0, 0,
		0, 15, 0, 104, 0, 48, 120, 56, 48, 48, 48, 55, 48, 54, 98, 0, 0, 48, 0, 101,
		0, 16, 0, 103, 0, 102, 119, 46, 97, 112, 112, 46, 110, 97, 109, 101, 0, 27, 0,
		104, 0, 73, 67, 69, 32, 79, 83, 32, 68, 101, 102, 97, 117, 108, 116, 32, 80,
		97, 99, 107, 97, 103, 101, 0, 0, 32, 0, 101, 0, 11, 0, 103, 0, 102, 119, 46,
		97, 112, 112, 0, 0, 13, 0, 104, 0, 49, 46, 51, 46, 50, 52, 46, 48, 0, 0, 0, 0,
		44, 0, 101, 0, 21, 0, 103, 0, 102, 119, 46, 97, 112, 112, 46, 98, 117, 110,
		100, 108, 101, 95, 105, 100, 0, 0, 0, 0, 15, 0, 104, 0, 48, 120, 99, 48, 48,
		48, 48, 48, 48, 49, 0, 0, 44, 0, 101, 0, 15, 0, 103, 0, 102, 119, 46, 110,
		101, 116, 108, 105, 115, 116, 0, 0, 21, 0, 104, 0, 50, 46, 52, 48, 46, 50, 48,
		48, 48, 45, 51, 46, 49, 54, 46, 48, 0, 0, 0, 0, 44, 0, 101, 0, 21, 0, 103, 0,
		102, 119, 46, 110, 101, 116, 108, 105, 115, 116, 46, 98, 117, 105, 108, 100,
		0, 0, 0, 0, 15, 0, 104, 0, 48, 120, 54, 55, 54, 97, 52, 56, 57, 100, 0, 0}
}

func devlinkTestInfoParesd() map[string]string {
	// This data was obtained from Intel E810-C NIC
	return map[string]string{
		"board.id":         "K85585-000",
		"fw.app":           "1.3.24.0",
		"fw.app.bundle_id": "0xc0000001",
		"fw.app.name":      "ICE OS Default Package",
		"fw.bundle_id":     "0x8000706b",
		"fw.mgmt":          "5.4.5",
		"fw.mgmt.api":      "1.7",
		"fw.mgmt.build":    "0x391f7640",
		"fw.netlist":       "2.40.2000-3.16.0",
		"fw.netlist.build": "0x676a489d",
		"fw.psid.api":      "2.42",
		"fw.undi":          "1.2898.0",
	}
}
