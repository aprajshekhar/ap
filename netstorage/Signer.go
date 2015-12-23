// Signer.go
package netstorage

import (
	"github.com/fatih/structs"
)

type ApiEvent struct{
	Version				int
	Action				string
	AdditionalParams	map[string]string
	Format				string
	Destination			string
	Target				string
	TimeStamp			int
	Size				int64
	Md5					[]byte			
	Sha1				[]byte
	Sha256				[]byte
	IndexZip			bool		
}

func convertToMap(ApiEvent apiEvent)map[string]string{
	convertedMap := structs.Map(apiEvent)
	return convertedMap
}

