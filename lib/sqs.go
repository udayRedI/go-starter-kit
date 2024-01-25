package lib

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type AllowMessageDeleteError struct {
	message string
}

func (err AllowMessageDeleteError) Error() string {
	return err.message
}

func AllowQueueDeletionError(errMsg string) AllowMessageDeleteError {
	return AllowMessageDeleteError{
		message: errMsg,
	}
}

type ISqsManager interface {
	PublishToSQS(queueName string, messageBody string, requestId string) (string, error)
	HandleQueue(queueName *string, handler func(string, string, string) (func(string), error)) error
}

type SqsManager struct {
	sqsConnectoin ISqsConnection
}

type ISqsConnection interface {
	GetSession() *session.Session
	GetQueueUrl(queueName string) (string, error)
}

type AwsSqsConnection struct {
	Session *session.Session
}

func (sqsConn *AwsSqsConnection) GetSession() *session.Session {
	return sqsConn.Session
}
func (sqsConn *AwsSqsConnection) GetQueueUrl(queueName string) (string, error) {
	svc := sqs.New(sqsConn.Session)
	params := sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	}
	resp, err := svc.GetQueueUrl(&params)
	if err != nil {
		return "", err
	}
	return *resp.QueueUrl, nil
}

func newAwsSqsConnection() (*AwsSqsConnection, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_SECRET_REGION")),
	})

	return &AwsSqsConnection{
		sess,
	}, err
}

type LocalSqsConnection struct {
	Session *session.Session
}

func newLocalSqsConnection() *LocalSqsConnection {
	cfg := aws.Config{
		Region: aws.String(endpoints.UsEast1RegionID),
	}
	cfg.Endpoint = aws.String(os.Getenv("AWS_MOCK_ENDPOINT"))

	sess := session.Must(session.NewSession(&cfg))

	return &LocalSqsConnection{
		Session: sess,
	}
}

func (sqsConn *LocalSqsConnection) GetSession() *session.Session {
	return sqsConn.Session
}
func (sqsConn *LocalSqsConnection) GetQueueUrl(queueName string) (string, error) {
	url := fmt.Sprintf("%s%s", os.Getenv("AWS_MOCK_QUEUE_URL"), queueName)
	return url, nil
}

func NewSqsManager(env string) (*SqsManager, error) {

	var connection ISqsConnection

	if env == "LOCAL" {
		connection = newLocalSqsConnection()
	} else {
		_con, connErr := newAwsSqsConnection()
		if connErr != nil {
			return nil, connErr
		}
		connection = _con
	}

	sqsManager := SqsManager{
		sqsConnectoin: connection,
	}

	return &sqsManager, nil
}

func (sqsManager *SqsManager) getQueueURL(queueName string) (string, error) {
	return sqsManager.sqsConnectoin.GetQueueUrl(queueName)
}

func (sqsManager *SqsManager) PublishToSQS(queueName string, messageBody string, requestId string) (string, error) {
	log.Printf("DEBUG:%s Message body received in PublishSQS with request ID %s", messageBody, requestId)
	log.Printf("DEBUG:%s QueueName with request ID %s", queueName, requestId)
	urlRes, _ := sqsManager.getQueueURL(queueName)
	log.Printf("DEBUG: %s QueueURL with request ID %s", urlRes, requestId)
	sqsClient := sqs.New(sqsManager.sqsConnectoin.GetSession())

	resp, e := sqsClient.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    &urlRes,
		MessageBody: aws.String(messageBody),
	})
	if e != nil {
		log.Printf("%s Got an error while trying to send message to queue: %v", requestId, e)
		return "", e
	}
	log.Printf("INFO: Message sent successfully with request ID %s", requestId)
	return *resp.MessageId, nil
}

func (sqsManager *SqsManager) HandleQueue(queueName *string, handler func(string, string, string) (func(string), error)) error {
	sqsClient := sqs.New(sqsManager.sqsConnectoin.GetSession())
	urlRes, err := sqsManager.getQueueURL(*queueName)

	if err != nil {
		errTxt := fmt.Sprintf("Error %s initiating queue %s", err, *queueName)
		return errors.New(errTxt)
	}

	// Start a goroutine for handling messages
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		log.Printf("Successfully initiated queue %s", *queueName)
		wg.Done()
		for { // create an infinite processing loop
			requestId := GenerateRandomUUID()
			msgResult, recieveErr := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
				QueueUrl:              &urlRes,
				MaxNumberOfMessages:   aws.Int64(10),
				WaitTimeSeconds:       aws.Int64(1),
				MessageAttributeNames: aws.StringSlice([]string{"CustomData"}),
			})

			if recieveErr != nil {
				CaptureSentryException(fmt.Sprintf("Error: %s 'ReceiveMessage' from url(%s) function error: %s", requestId, *queueName, recieveErr))
				time.Sleep(30 * time.Second)
				continue
			}

			if len(msgResult.Messages) == 0 {
				time.Sleep(1 * time.Second)
				continue
			}

			go func(messages []*sqs.Message) {
				for i := range messages {
					go func(itr int) {
						defer Handlepanic(fmt.Sprintf("%s: Error running queue(%s)", requestId, *queueName))
						message := msgResult.Messages[itr]
						customData := ""
						if message.MessageAttributes != nil {
							if customDataAttr, ok := message.MessageAttributes["CustomData"]; ok {
								customData = *customDataAttr.StringValue
							}
						}

						body := message.Body
						receiptHandle := message.ReceiptHandle

						cleanup, handlerErr := handler(*body, customData, requestId)

						if _, works := handlerErr.(AllowMessageDeleteError); handlerErr != nil && !works {
							CaptureSentryException(fmt.Sprintf("%s Failed to process message on queue(%s) with error %s", requestId, *queueName, handlerErr.Error()))
							log.Printf("%s Skipping message delete", requestId)
							return
						}

						_, deleteErr := sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
							QueueUrl:      &urlRes,
							ReceiptHandle: receiptHandle,
						})

						if deleteErr != nil {
							log.Printf("Error: DeleteMessage error %s for queue(%s) with request ID %s", err, *queueName, requestId)
						}

						cleanup(requestId)
					}(i)
				}
			}(msgResult.Messages)
		}
	}()
	wg.Wait()

	// Return immediately, leaving the goroutine running in the background
	return nil
}
