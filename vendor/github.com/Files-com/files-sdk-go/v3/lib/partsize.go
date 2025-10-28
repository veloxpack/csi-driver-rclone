package lib

const (
	BasePart = int64(1024 * 1024 * 5)
)

// PartSizes return 10k parts that add up to 4.9 TB
var PartSizes = partSizes()

func partSizes() []int64 {
	maxSize := int64(1024 * 1024 * 1024 * 1024 * 5)
	totalParts := 10_000
	totalSize := int64(0)
	var parts []int64
	partSize := BasePart
	var iter []int
	iter = append(iter, 0)
	iter = append(iter, 100)
	f := fibonacci()
	f() // 0 - skip first two rounds of fib
	f() // 100
	for len(parts) < totalParts {
		perIt := f() * 100
		for i := 0; i < perIt; i++ {
			if len(parts) == totalParts {
				break
			}

			remaining := maxSize - totalSize
			if remaining >= partSize {
				parts = append(parts, partSize)
				totalSize += partSize
			} else if remaining > 0 {
				parts = append(parts, remaining)
				totalSize += remaining
			} else {
				break
			}
		}

		if maxSize-totalSize == 0 {
			break
		}
		partSize = partSize + partSize
	}
	return parts
}

func fibonacci() func() int {
	first, second := 0, 1
	return func() int {
		ret := first
		first, second = second, first+second
		return ret
	}
}
