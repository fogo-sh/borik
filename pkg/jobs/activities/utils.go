package activities

import "github.com/mitchellh/mapstructure"

func decodeOperationArgs(args OperationArgs, targetPtr any) error {
	// On Temporal decoding, the args come through as a map[string]any, rather than our desired type
	mapStruct := args.Args.(map[string]any)

	return mapstructure.Decode(mapStruct, targetPtr)
}
