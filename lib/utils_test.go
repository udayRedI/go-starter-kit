package lib

import "testing"

func TestChunkArray(t *testing.T) {

	type TestCase struct {
		Title string
		Size  int
		Input []string
		Exp   [][]string
		Got   [][]string
	}

	testCases := []TestCase{
		{
			Title: "Array of 6 and chunk size 3",
			Size:  3,
			Input: []string{
				"One",
				"Two",
				"Three",
				"Four",
				"Five",
				"Six",
			},
			Exp: [][]string{
				[]string{
					"One",
					"Two",
					"Three",
				},
				[]string{
					"Four",
					"Five",
					"Six",
				},
			},
		},
		{
			Title: "Array of 6 and chunk size 5",
			Size:  5,
			Input: []string{
				"One",
				"Two",
				"Three",
				"Four",
				"Five",
				"Six",
			},
			Exp: [][]string{
				[]string{
					"One",
					"Two",
					"Three",
					"Four",
					"Five",
				},
				[]string{
					"Six",
				},
			},
		},
		{
			Title: "Array of 1 and chunk size 10",
			Size:  10,
			Input: []string{
				"One",
			},
			Exp: [][]string{
				[]string{
					"One",
				},
			},
		},
		{
			Title: "Array of 6 and chunk size 6",
			Size:  6,
			Input: []string{
				"One",
				"Two",
				"Three",
				"Four",
				"Five",
				"Six",
			},
			Exp: [][]string{
				[]string{
					"One",
					"Two",
					"Three",
					"Four",
					"Five",
					"Six",
				},
			},
		},
		{
			Title: "Array of 6 and chunk size 6",
			Size:  1,
			Input: []string{
				"One",
				"Two",
				"Three",
				"Four",
				"Five",
				"Six",
			},
			Exp: [][]string{
				[]string{
					"One",
				},
				[]string{
					"Two",
				},
				[]string{
					"Three",
				},
				[]string{
					"Four",
				},
				[]string{
					"Five",
				},
				[]string{
					"Six",
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Title, func(t *testing.T) {
			testCase.Got = ChunkArray(testCase.Input, testCase.Size)
			if len(testCase.Exp) != len(testCase.Got) {
				t.Errorf("Returned array is different from what was expected")
			}
			for outerInd, expItemArr := range testCase.Exp {
				for innerInd, item := range expItemArr {
					if item != testCase.Got[outerInd][innerInd] {
						t.Errorf("Returned array is different from what was expected")
					}
				}
			}
		})
	}
}
