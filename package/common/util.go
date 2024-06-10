package common

func AddLB(data []byte) []byte {
	return append(data, '\n')
}
