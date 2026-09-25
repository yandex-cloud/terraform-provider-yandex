package topicreadercommon

import "github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic/rawtopicreader"

func ReadRawBatchesToPublicBatches(
	msg *rawtopicreader.ReadResponse,
	sessions *PartitionSessionStorage,
	decoders DecoderMap,
) ([]*PublicBatch, error) {
	batchesCount := 0
	for i := range msg.PartitionData {
		batchesCount += len(msg.PartitionData[i].Batches)
	}

	var batches []*PublicBatch
	for pIndex := range msg.PartitionData {
		p := &msg.PartitionData[pIndex]

		// normal way
		session, err := sessions.Get(p.PartitionSessionID)
		if err != nil {
			return nil, err
		}

		for bIndex := range p.Batches {
			batch, err := NewBatchFromStream(decoders, session, p.Batches[bIndex])
			if err != nil {
				return nil, err
			}
			batches = append(batches, batch)
		}
	}

	if err := splitBytesByMessagesInBatches(batches, msg.BytesSize); err != nil {
		return nil, err
	}

	return batches, nil
}
