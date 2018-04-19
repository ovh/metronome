package core_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/ovh/metronome/src/metronome/models"

	core "github.com/ovh/metronome/src/scheduler/core"
)

func entry(schedule string) (*core.Entry, error) {
	return core.NewEntry(models.Task{
		Schedule: schedule,
	})
}

var _ = Describe("Entry", func() {
	Describe("New", func() {
		It("Good schedule", func() {
			_, err := entry("R/2016-12-15T11:39:00Z/PT1S/ET1S")
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("Same as", func() {
			task := models.Task{
				UserID:   "UserID",
				GUID:     "GUID",
				URN:      "URN",
				Schedule: "R/2016-12-15T11:39:00Z/PT1S/ET1S",
			}

			entry, err := core.NewEntry(task)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(entry.SameAs(task)).Should(BeTrue())
			Ω(entry.UserID() == task.UserID).Should(BeTrue())
			Ω(entry.URN() == task.URN).Should(BeTrue())
			Ω(entry.GUID() == task.GUID).Should(BeTrue())
		})

		It("Bad schedule", func() {
			_, err := entry("BadSchedule")
			Ω(err).Should(HaveOccurred())
		})

		It("Bad schedule date", func() {
			_, err := entry("R/notAdateZ/PT1S/ET1S")
			Ω(err).Should(HaveOccurred())
		})

		It("Null time period", func() {
			_, err := entry("R/2016-12-15T11:39:00Z/PT0S/ET1S")
			Ω(err).Should(HaveOccurred())
		})

		It("Null date period", func() {
			_, err := entry("R/2016-12-15T11:39:00Z/P0M/ET1S")
			Ω(err).Should(HaveOccurred())
		})

		It("With repeat", func() {
			_, err := entry("R3/2016-12-15T11:39:00Z/P1M/ET1S")
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("Bad repeat", func() {
			_, err := entry("Ra/2016-12-15T11:39:00Z/P1M/ET1S")
			Ω(err).Should(HaveOccurred())
		})
	})

	DescribeTable("Init",
		func(schedule, now string, next string) {
			entry, err := entry(schedule)
			Ω(err).ShouldNot(HaveOccurred())

			n, err := time.Parse(time.RFC3339, now)
			Ω(err).ShouldNot(HaveOccurred())

			entry.Init(n)

			if len(next) > 0 {
				nx, err := time.Parse(time.RFC3339, next)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(time.Unix(entry.Next(), 0).UTC()).Should(BeTemporally("==", nx))
			} else {
				Ω(entry.Next()).Should(Equal(int64(-1)))
			}
		},
		// Time
		Entry("in the future", "R/2016-12-15T11:32:00Z/PT1S/ET1S", "2016-01-01T00:00:00Z", "2016-12-15T11:32:00Z"),
		Entry("1s", "R/2016-12-01T00:00:00Z/PT1S/ET1S", "2017-01-01T00:00:00Z", "2017-01-01T00:00:00Z"),
		Entry("10s start time", "R/2017-01-01T00:00:00Z/PT10S/ET1S", "2017-01-01T00:00:00Z", "2017-01-01T00:00:00Z"),
		Entry("10s on time", "R/2016-01-01T00:00:00Z/PT10S/ET1S", "2017-01-01T00:00:00Z", "2017-01-01T00:00:00Z"),
		Entry("10s", "R/2017-01-01T00:00:00Z/PT10S/ET1S", "2017-01-01T00:00:03Z", "2017-01-01T00:00:10Z"),
		Entry("1m", "R/2017-01-01T00:00:00Z/PT1M/ET1S", "2017-01-01T00:00:03Z", "2017-01-01T00:01:00Z"),
		Entry("1m10s", "R/2017-01-01T00:00:00Z/PT1M10S/ET1S", "2017-01-01T00:00:03Z", "2017-01-01T00:01:10Z"),
		Entry("1h", "R/2017-01-01T00:00:00Z/PT1H/ET1S", "2017-01-01T00:00:03Z", "2017-01-01T01:00:00Z"),
		Entry("1h5m24s", "R/2017-01-01T00:00:00Z/PT1H5M24S/ET1S", "2017-01-01T00:00:03Z", "2017-01-01T01:05:24Z"),
		Entry("1D", "R/2017-01-01T00:00:00Z/P1DT/ET1S", "2017-01-01T00:00:03Z", "2017-01-02T00:00:00Z"),
		Entry("1D3h27m17s", "R/2017-01-01T00:00:00Z/P1DT3H27M17S/ET1S", "2017-01-01T00:00:03Z", "2017-01-02T03:27:17Z"),
		// Date
		Entry("in the future", "R/2016-12-15T11:32:00Z/P1M/ET1S", "2016-01-01T00:00:00Z", "2016-12-15T11:32:00Z"),
		Entry("1M start time", "R/2017-01-01T00:00:00Z/P1M/ET1S", "2017-01-01T00:00:00Z", "2017-01-01T00:00:00Z"),
		Entry("1M on time", "R/2016-01-01T00:00:00Z/P1M/ET1S", "2017-01-01T00:00:00Z", "2017-01-01T00:00:00Z"),
		Entry("1M", "R/2017-01-01T00:00:00Z/P1M/ET1S", "2017-01-01T00:00:03Z", "2017-02-01T00:00:00Z"),
		Entry("1M end month", "R/2017-01-31T00:00:00Z/P1M/ET1S", "2017-02-01T00:00:03Z", "2017-02-28T00:00:00Z"),
		Entry("3M", "R/2017-01-01T00:00:00Z/P3M/ET1S", "2017-01-01T00:00:03Z", "2017-04-01T00:00:00Z"),
		Entry("3M end month", "R/2017-01-30T00:00:00Z/P3M/ET1S", "2017-01-31T00:00:03Z", "2017-04-30T00:00:00Z"),
		Entry("1Y", "R/2017-01-03T00:00:00Z/P1Y/ET1S", "2017-01-31T00:00:03Z", "2018-01-03T00:00:00Z"),
		Entry("1Y5M", "R/2017-01-03T03:00:00Z/P1Y5M/ET1S", "2017-01-31T00:00:03Z", "2018-06-03T03:00:00Z"),
		// Repeat
		Entry("10s no repeat", "R0/2017-01-01T00:00:00Z/PT10S/ET1S", "2017-01-01T00:00:00Z", "2017-01-01T00:00:00Z"),
		Entry("10s no repeat over", "R0/2017-01-01T00:00:00Z/PT10S/ET1S", "2017-01-01T00:00:01Z", ""),
		Entry("10s R3", "R3/2017-01-01T00:00:00Z/PT10S/ET1S", "2017-01-01T00:00:30Z", "2017-01-01T00:00:30Z"),
		Entry("10s R3 over", "R3/2017-01-01T00:00:00Z/PT10S/ET1S", "2017-01-01T00:00:31Z", ""),
		Entry("5M no repeat", "R0/2017-01-01T00:00:00Z/P5M/ET1S", "2017-01-01T00:00:00Z", "2017-01-01T00:00:00Z"),
		Entry("5M no repeat over", "R0/2017-01-01T00:00:00Z/P5M/ET1S", "2017-01-01T00:00:01Z", ""),
		Entry("5M R1", "R1/2017-01-01T00:00:00Z/P5M/ET1S", "2017-06-01T00:00:00Z", "2017-06-01T00:00:00Z"),
		Entry("5M R1 over", "R1/2017-01-01T00:00:00Z/P5M/ET1S", "2017-06-01T00:00:01Z", ""),
	)

	Describe("Plan", func() {
		It("Should return an error if not initialized", func() {
			entry, err := entry("R/2016-12-15T11:39:00Z/PT1S/ET1S")
			Ω(err).ShouldNot(HaveOccurred())

			_, err = entry.Plan(time.Now())
			Ω(err).Should(HaveOccurred())
		})

		DescribeTable("Next",
			func(schedule string, plans []struct {
				now  string
				next string
			}) {
				entry, err := entry(schedule)
				Ω(err).ShouldNot(HaveOccurred())

				for i, plan := range plans {
					By(fmt.Sprintf("n%v - %v", i, plan.now))

					n, err := time.Parse(time.RFC3339, plans[i].now)
					Ω(err).ShouldNot(HaveOccurred())
					if i == 0 {
						entry.Init(n)
					} else {
						entry.Plan(n)
					}

					if len(plan.next) > 0 {
						nx, err := time.Parse(time.RFC3339, plan.next)
						Ω(err).ShouldNot(HaveOccurred())

						Ω(time.Unix(entry.Next(), 0).UTC()).Should(BeTemporally("==", nx))
					} else {
						Ω(entry.Next()).Should(Equal(int64(-1)))
					}

				}
			},
			// Time
			Entry("in the future", "R/2016-12-15T11:32:00Z/PT1S/ET1S", []struct {
				now  string
				next string
			}{{"2016-01-01T00:00:00Z", "2016-12-15T11:32:00Z"},
				{"2016-01-01T00:00:01Z", "2016-12-15T11:32:00Z"}}),
			Entry("10s", "R/2017-01-01T00:00:00Z/PT10S/ET1S", []struct {
				now  string
				next string
			}{{"2017-01-01T00:00:03Z", "2017-01-01T00:00:10Z"},
				{"2017-01-01T00:00:11Z", "2017-01-01T00:00:20Z"},
				{"2017-01-01T00:00:12Z", "2017-01-01T00:00:20Z"},
				{"2017-01-01T00:00:20Z", "2017-01-01T00:00:20Z"},
				{"2017-01-01T00:00:21Z", "2017-01-01T00:00:30Z"}}),
			// date
			Entry("3M", "R/2017-01-01T00:00:00Z/P3M/ET1S", []struct {
				now  string
				next string
			}{{"2017-01-01T00:00:00Z", "2017-01-01T00:00:00Z"},
				{"2017-01-01T00:00:11Z", "2017-04-01T00:00:00Z"},
				{"2017-01-02T00:00:12Z", "2017-04-01T00:00:00Z"},
				{"2017-02-01T00:00:00Z", "2017-04-01T00:00:00Z"},
				{"2017-04-01T00:00:21Z", "2017-07-01T00:00:00Z"}}),
			Entry("1M end month", "R/2017-01-31T00:00:00Z/P1M/ET1S", []struct {
				now  string
				next string
			}{{"2017-01-01T00:00:00Z", "2017-01-31T00:00:00Z"},
				{"2017-01-31T00:00:00Z", "2017-01-31T00:00:00Z"},
				{"2017-01-31T01:00:00Z", "2017-02-28T00:00:00Z"},
				{"2017-02-15T00:00:00Z", "2017-02-28T00:00:00Z"},
				{"2017-02-28T01:00:00Z", "2017-03-31T00:00:00Z"}}),
			// repeat
			Entry("7m repeat start time", "R2/2017-01-01T00:00:00Z/PT7M/ET1S", []struct {
				now  string
				next string
			}{{"2017-01-01T00:00:00Z", "2017-01-01T00:00:00Z"},
				{"2017-01-01T00:05:00Z", "2017-01-01T00:07:00Z"},
				{"2017-01-01T00:08:00Z", "2017-01-01T00:14:00Z"},
				{"2017-01-01T00:14:01Z", ""}}),
			Entry("2h repeat", "R1/2017-01-01T00:00:00Z/PT2H/ET1S", []struct {
				now  string
				next string
			}{{"2017-01-01T00:00:03Z", "2017-01-01T02:00:00Z"},
				{"2017-01-01T01:30:00Z", "2017-01-01T02:00:00Z"},
				{"2017-01-01T02:00:00Z", "2017-01-01T02:00:00Z"},
				{"2017-01-01T02:14:01Z", ""}}),
		)
	})
})
