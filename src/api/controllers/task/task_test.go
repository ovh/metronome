package taskctrl

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Task", func() {
	table.DescribeTable("Validate_creation", func(name, schedule string, shouldFail bool) {
		err := validateCreation(schedule)
		if shouldFail {
			Ω(err).Should(HaveOccurred())
		} else {
			Ω(err).ShouldNot(HaveOccurred())
		}
	},
		table.Entry("12 h of period and 6 h of epsilon", "R/2018-11-22T10:12:01Z/PT12H/ET6H", false),
		table.Entry("6 min of period and 3 min of epsilon", "R/2018-11-21T14:49:11Z/PT6M/ET3M", false),
		table.Entry("2 min of period 1 m of epsilon", "R/2018-11-15T18:06:47Z/PT2M/ET1M", false),
		table.Entry("1 day and 3 seconds of epsilon", "R/2016-12-31T10:00:00Z/P1DT/ET3S", false),
		table.Entry("DHMS 1", "R/2016-12-31T10:00:00Z/P31DT1H5M32S/ET3S", false),
		table.Entry("DHMS 2", "R/2016-12-31T10:00:00Z/P31DT1H5M/ET3S", false),
		table.Entry("DHM 1", "R/2016-12-31T10:00:00Z/P31DT1H5M/ET3S", false),
		table.Entry("DHM 2 (failed because no value before M)", "R/2016-12-31T10:00:00Z/P31DT1HM32S/ET3S", true),
		table.Entry("DMS 1", "R/2016-12-31T10:00:00Z/P31DT5M32S/ET3S", false),
		table.Entry("DMS 2 (failed because no value before H)", "R/2016-12-31T10:00:00Z/P31DTH5M32S/ET3S", true),
		table.Entry("HMS 1", "R/2016-12-31T10:00:00Z/PT1H5M32S/ET3S", true),
		table.Entry("HMS 2 (missing T)", "R/2016-12-31T10:00:00Z/P31T1H5M32S/ET3S", true),
		table.Entry("DH 1", "R/2016-12-31T10:00:00Z/P31DT1H/ET3S", true),
		table.Entry("DH 2 (No unit after 5)", "R/2016-12-31T10:00:00Z/P31DT1H5/ET3S", true),
		table.Entry("DM 1", "R/2016-12-31T10:00:00Z/P31DT5M/ET3S", true),
		table.Entry("DM 2 (No unit after 32)", "R/2016-12-31T10:00:00Z/P31DT5M32/ET3S", true),
		table.Entry("DS 1", "R/2016-12-31T10:00:00Z/P31DT32S/ET3S", true),
		table.Entry("DS 2 (No H value)", "R/2016-12-31T10:00:00Z/P31DTH32S/ET3S", true),
		table.Entry("HM 1", "R/2016-12-31T10:00:00Z/PT1H5M/ET3S", true),
		table.Entry("HM 2 (No unit after 32)", "R/2016-12-31T10:00:00Z/PT1H5M32/ET3S", true),
		table.Entry("HS 1", "R/2016-12-31T10:00:00Z/PT1H32S/ET3S", true),
		table.Entry("HS 2 (No value before M)", "R/2016-12-31T10:00:00Z/PT1HM32S/ET3S", true),
		table.Entry("MS 1", "R/2016-12-31T10:00:00Z/PT5M32S/ET3S", true),
		table.Entry("MS 2 (No value before H)", "R/2016-12-31T10:00:00Z/PTH5M32S/ET3S", true),
		table.Entry("D 1", "R/2016-12-31T10:00:00Z/P1DT/ET3S", true),
		table.Entry("D 2 (No unit after 5)", "R/2016-12-31T10:00:00Z/PT1H5/ET3S", true),
		table.Entry("H 1", "R/2016-12-31T10:00:00Z/PT1H/ET3S", false),
		table.Entry("H 2 (No unit after 5)", "R/2016-12-31T10:00:00Z/PT1H5/ET3S", false),
		table.Entry("M 1", "R/2016-12-31T10:00:00Z/PT5M/ET3S", false),
		table.Entry("M 2 (No unit after 32)", "R/2016-12-31T10:00:00Z/PT5M32/ET3S", false),
		table.Entry("S 1", "R/2016-12-31T10:00:00Z/PT32S/ET3S", false),
		table.Entry("S 2 (No unit after 32)", "R/2016-12-31T10:00:00Z/PT32/ETM3S", false),
	)
})
