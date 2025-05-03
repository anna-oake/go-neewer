package protocol

import "errors"

// 8-bit sum mod 256
func computeChecksum(data []byte) byte {
	var sum byte
	for _, b := range data {
		sum += b
	}
	return sum
}

// will split payload in format [length][data][length][data]... where [length] is two bytes
func decodeLV(data []byte) ([][]byte, error) {
	if len(data) < 2 {
		return nil, errors.New("data too short")
	}

	var parcels [][]byte
	for len(data) > 0 {
		length := int(data[0])<<8 | int(data[1])
		if length > len(data)-2 {
			return nil, errors.New("data too short for length")
		}
		parcels = append(parcels, data[2:2+length])
		data = data[2+length:]
	}
	return parcels, nil
}

func encodeLV(data [][]byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("no data to encode")
	}

	var result []byte
	for _, parcel := range data {
		if len(parcel) > 65535 {
			return nil, errors.New("parcel too large")
		}
		result = append(result, byte(len(parcel)>>8), byte(len(parcel)))
		result = append(result, parcel...)
	}
	return result, nil
}
