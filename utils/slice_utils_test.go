package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"strconv"
)

var _ = Describe("using slice utils", func() {
	When("calling mapSlice", func() {
		It("should return a slice applying a 'map' function", func() {
			// convert a slice of string to a slice of numbers with the custom map function
			slice := []string{"1", "2", "3"}
			mappedSlice := MapSlice(slice, func(stringVal string) int64 {
				res, _ := strconv.ParseInt(stringVal, 10, 0)
				return res
			})
			Expect(slice).To(HaveLen(len(mappedSlice)))
			//Expect(found).To(BeTrue())
		})
	})
})
