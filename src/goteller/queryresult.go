package goteller

import (
	"../ipaddr"
	"../messages"
)

type HitResult messages.HitResult

type QueryResult struct {
	fileIndex uint32
	fileSize  uint32
	filename  string
	addr      ipaddr.IPAddr
}

func (qr *QueryResult) GetFileIndex() uint32 {
	return qr.fileIndex
}

func (qr *QueryResult) GetFileSize() uint32 {
	return qr.fileSize
}

func (qr *QueryResult) GetFilename() string {
	return qr.filename
}

func resultsFromHit(queryHit messages.QueryHitMsg) []QueryResult {
	numResults := len(queryHit.ResultSet)
	if numResults == 0 {
		return []QueryResult{} // Empty slice
	}
	results := make([]QueryResult, numResults)
	for i, hit := range queryHit.ResultSet {
		results[i] = QueryResult{
			fileIndex: hit.FileIndex,
			fileSize:  hit.FileSize,
			filename:  hit.Filename,
			addr:      queryHit.Addr,
		}
	}
	return results
}
