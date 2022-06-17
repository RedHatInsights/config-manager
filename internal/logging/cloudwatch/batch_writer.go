package cloudwatch

import (
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

// BatchWriter is an io.Writer that batch writes log events to the AWS
// CloudWatch logs API.
type BatchWriter struct {
	svc               *cloudwatchlogs.CloudWatchLogs
	groupName         string
	streamName        string
	nextSequenceToken *string
	m                 sync.Mutex
	ch                chan *cloudwatchlogs.InputLogEvent
	flushWG           sync.WaitGroup
	err               *error
}

// NewBatchWriter creates a BatchWriter with the given group and stream. To
// create an unbuffered writer, set batchFrequency to 0.
func NewBatchWriter(groupName, streamName string, cfg *aws.Config, batchFrequency time.Duration) (*BatchWriter, error) {
	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}
	w := &BatchWriter{
		svc:        cloudwatchlogs.New(sess),
		groupName:  groupName,
		streamName: streamName,
	}

	resp, err := w.getOrCreateCloudWatchLogGroup()
	if err != nil {
		return nil, err
	}

	if batchFrequency > 0 {
		w.ch = make(chan *cloudwatchlogs.InputLogEvent, 10000)
		ticker := time.NewTicker(batchFrequency)

		go w.putBatches(ticker.C)
	}

	// grab the next sequence token
	if len(resp.LogStreams) > 0 {
		w.nextSequenceToken = resp.LogStreams[0].UploadSequenceToken
		return w, nil
	}

	// create stream if it doesn't exist. the next sequence token will be null
	_, err = w.svc.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(groupName),
		LogStreamName: aws.String(streamName),
	})
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (w *BatchWriter) getOrCreateCloudWatchLogGroup() (*cloudwatchlogs.DescribeLogStreamsOutput, error) {
	resp, err := w.svc.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(w.groupName),
		LogStreamNamePrefix: aws.String(w.streamName),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case cloudwatchlogs.ErrCodeResourceNotFoundException:
				_, err = w.svc.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
					LogGroupName: aws.String(w.groupName),
				})
				if err != nil {
					return nil, err
				}
				return w.getOrCreateCloudWatchLogGroup()
			default:
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return resp, nil
}

func (w *BatchWriter) putBatches(ticker <-chan time.Time) {
	var batch []*cloudwatchlogs.InputLogEvent
	size := 0
	for {
		select {
		case p := <-w.ch:
			if p != nil {
				messageSize := len(*p.Message) + 26
				if size+messageSize >= 1048576 || len(batch) == 10000 {
					w.sendBatch(batch)
					batch = nil
					size = 0
				}
				batch = append(batch, p)
				size += messageSize
			} else {
				// Flush event (nil)
				w.sendBatch(batch)
				w.flushWG.Done()
				batch = nil
				size = 0
			}
		case <-ticker:
			w.sendBatch(batch)
			batch = nil
			size = 0
		}
	}
}

func (w *BatchWriter) sendBatch(batch []*cloudwatchlogs.InputLogEvent) {
	if len(batch) == 0 {
		return
	}
	params := &cloudwatchlogs.PutLogEventsInput{
		LogEvents:     batch,
		LogGroupName:  aws.String(w.groupName),
		LogStreamName: aws.String(w.streamName),
		SequenceToken: w.nextSequenceToken,
	}
	resp, err := w.svc.PutLogEvents(params)
	if err == nil {
		w.nextSequenceToken = resp.NextSequenceToken
		return
	}

	w.err = &err
	if aerr, ok := err.(*cloudwatchlogs.InvalidSequenceTokenException); ok {
		w.nextSequenceToken = aerr.ExpectedSequenceToken
		w.sendBatch(batch)
		return
	}
}

// Flush immediately writes any pending events to the CloudWatch API.
func (w *BatchWriter) Flush() error {
	w.flushWG.Add(1)
	w.ch <- nil
	w.flushWG.Wait()
	if w.err != nil {
		return *w.err
	}
	return nil
}

func (w *BatchWriter) Write(p []byte) (n int, err error) {
	event := &cloudwatchlogs.InputLogEvent{
		Message:   aws.String(string(p)),
		Timestamp: aws.Int64(int64(time.Nanosecond) * time.Now().UnixNano() / int64(time.Millisecond)),
	}

	if w.ch != nil {
		w.ch <- event
		if w.err != nil {
			lastErr := w.err
			w.err = nil
			return 0, fmt.Errorf("%v", *lastErr)
		}
		return len(p), nil
	}

	w.m.Lock()
	defer w.m.Unlock()

	params := &cloudwatchlogs.PutLogEventsInput{
		LogEvents:     []*cloudwatchlogs.InputLogEvent{event},
		LogGroupName:  aws.String(w.groupName),
		LogStreamName: aws.String(w.streamName),
		SequenceToken: w.nextSequenceToken,
	}
	resp, err := w.svc.PutLogEvents(params)
	if err != nil {
		return 0, err
	}

	w.nextSequenceToken = resp.NextSequenceToken

	return len(p), nil
}
