package utils

type Upload struct {
	Name    string
	Process string
}

func NewUpload(name, process string) Upload {
	return Upload{
		name,
		process,
	}
}

func ProcessBatch(tasks []Upload, batch int) [][]Upload {
	pool := [][]Upload{}

	for i := 0; i < len(tasks); i += batch {
		end := i + batch

		if end > len(tasks) {
			end = len(tasks)
		}

		pool = append(pool, tasks[i:end])
	}

	return pool
}

/* data
tasks1 := []Upload{
	NewUpload("1", "1"),
	NewUpload("2", "2"),
	NewUpload("3", "3"),
	NewUpload("4", "4"),

	NewUpload("5", "5"),
	NewUpload("6", "6"),
	NewUpload("7", "7"),
	NewUpload("8", "8"),

	NewUpload("9", "9"),
	NewUpload("10", "10"),
	NewUpload("11", "11"),
	NewUpload("12", "12"),

	NewUpload("13", "13"),
	NewUpload("14", "14"),
	NewUpload("15", "15"),

	NewUpload("16", "16"),
	NewUpload("17", "17"),
	NewUpload("18", "18"),

	NewUpload("19", "19"),
	NewUpload("20", "20"),
	NewUpload("21", "21"),

	NewUpload("22", "22"),
	NewUpload("23", "23"),
	NewUpload("24", "24"),

	NewUpload("25", "25"),
	NewUpload("26", "26"),
	NewUpload("27", "27"),
	NewUpload("28", "28"),

	NewUpload("29", "29"),
	NewUpload("30", "30"),
	NewUpload("31", "31"),
	NewUpload("32", "32"),

	NewUpload("33", "33"),
	NewUpload("34", "34"),
	NewUpload("35", "35"),
}

*/
