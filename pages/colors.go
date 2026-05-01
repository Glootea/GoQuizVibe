package pages

import "fmt"

func PositionClass(position int) string {
	switch position {
	case 1:
		return "bg-yellow-400 text-yellow-900"
	case 2:
		return "bg-gray-300 text-gray-700"
	case 3:
		return "bg-orange-400 text-orange-900"
	default:
		return "bg-gray-100 text-gray-600"
	}
}

func RowClass(position int) string {
	if position <= 3 {
		return fmt.Sprintf("flex items-center gap-4 p-4 rounded-xl hover:bg-gray-50 transition %s",
			"bg-gradient-to-r from-yellow-50 to-orange-50 border border-yellow-200")
	}
	return "flex items-center gap-4 p-4 rounded-xl hover:bg-gray-50 transition"
}
