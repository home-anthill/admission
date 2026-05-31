package utils

import (
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

	When("calling Filter", func() {
		It("should return only the elements matching the predicate", func() {
			filtered := Filter([]int{1, 2, 3, 4}, func(v int) bool {
				return v%2 == 0
			})

			Expect(filtered).To(Equal([]int{2, 4}))
		})
	})
})
