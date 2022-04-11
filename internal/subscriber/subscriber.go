package subscriber

import (
	"cloud.google.com/go/pubsub"
	pubsubClient "cloud.google.com/go/pubsub"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/harisbeha/media-transcoder/internal/actions"
	config "github.com/harisbeha/media-transcoder/internal/config"
	"github.com/harisbeha/media-transcoder/internal/data"
	models "github.com/harisbeha/media-transcoder/internal/models"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"math/rand"
	"sync"
	"time"
)

var (
	redisPool *redis.Pool
	downloadEnqueuer  *work.Enqueuer
	transcodeEnqueuer *work.Enqueuer
)

// Config defines configuration for creating a NewServer.
type Config struct {
	ServerPort  string
	RedisHost   string
	RedisPort   int
	Namespace   string
	JobName     string
	Concurrency uint
}

type request struct {
	C24JobId    string `json:"c24_job_id" binding:"required"`
	Profile     string `json:"profile" binding:"required"`
	Source      string `json:"source" binding:"required"`
	Destination string `json:"dest" binding:"required"`
	Action      string `json:"action" binding:"action"`
}

type updateRequest struct {
	Status string `json:"status"`
}

type response struct {
	Message string      `json:"message"`
	Status  int         `json:"status"`
	Job     *models.Job `json:"job"`
}

type index struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Docs    string `json:"docs"`
	Github  string `json:"github"`
}

type machineRequest struct {
	Provider string `json:"provider" binding:"required"`
	Region   string `json:"region" binding:"required"`
	Size     string `json:"size" binding:"required"`
	Count    int    `json:"count" binding:"required,min=1,max=10"` // Max of 10 machines.
}

type PushRequest struct {
	Message struct {
		Attributes map[string]string
		Data       []byte
		ID         string `json:"message_id"`
	}
	Subscription string
}

type Data struct {
	Key  string                 `json:"key"`
	Body string                 `json:"body"`
	Meta map[string]interface{} `json:"meta"`
}

type Worker struct {
	taskrAddr        string
	projectID        string
	topicName        string
	subscriptionName string
	client           *pubsubClient.Client
	pubsub           *pubsub.Client
	topic            *pubsub.Topic
	subscription     *pubsub.Subscription
	messages         chan *pubsub.Message
}

// SubscribeMessageHandler that handles the message
type SubscribeMessageHandler func(chan *pubsub.Message)

// ErrorHandler that logs the error received while reading a message
type ErrorHandler func(error)

// SubscriberConfig subscriber config
type SubscriberConfig struct {
	ProjectID        string
	TopicName        string
	SubscriptionName string
	ErrorHandler     ErrorHandler
	Handle           SubscribeMessageHandler
}

// Subscriber subscribe to a topic and pass each message to the
// handler function
type Subscriber struct {
	topic        *pubsub.Topic
	subscription *pubsub.Subscription
	errorHandler ErrorHandler
	handle       SubscribeMessageHandler
	cancel       func()
}

type sampleMsg struct {
	EventID string `json:"event_id"`
}

// NewServer creates a new server
func NewSubscriber(serverCfg Config) {
	ctx := context.Background()

	rand.Seed(time.Now().UnixNano())

	if config.PubsubClient == nil {
		log.Fatal("You must configure the Pub/Sub client in config.go before running pubsub_worker.")
	}
	// Setup redis queue.
	redisPool = &redis.Pool{
		MaxActive: 5,
		MaxIdle:   5,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(fmt.Sprintf("%s:%d", serverCfg.RedisHost, serverCfg.RedisPort))
		},
	}

	 downloadEnqueuer = work.NewEnqueuer("download", redisPool)
	 transcodeEnqueuer = work.NewEnqueuer("transcode", redisPool)

	//downloadEnqueuer = work.NewEnqueuer(serverCfg.Namespace, redisPool)
	//transcodeEnqueuer = work.NewEnqueuer(serverCfg.Namespace, redisPool)

	// Sets your Google Cloud Platform project ID.
	projectID := "coresystem-171219"

	subscriberConfig := SubscriberConfig{
		ProjectID:        projectID,
		TopicName:        config.PubsubTopicID,
		SubscriptionName: config.PubsubTopicSubscription,
		ErrorHandler: func(err error) {
			log.Printf("Subscriber error: %v", err)
		},
		Handle: func(output chan *pubsub.Message) {
			for {
				pMsg := <-output
				log.Printf("Message: %+v", string(pMsg.Data))
				newMsg := &request{}
				if err := json.Unmarshal(pMsg.Data, &newMsg); err != nil {
					log.Errorf("failed to unmarshal message body", err)
					return
				}
				go CreateJob(*newMsg)
				log.Info("profile:", newMsg.Profile, "source", newMsg.Source, "dest:", newMsg.Destination)
				pMsg.Ack()
			}
		},
	}
	log.Printf("Subscriber config: %+v", subscriberConfig)
	subscriber, err := CreateSubscription(ctx, subscriberConfig)
	if err != nil {
		log.Fatalf("Error occured while creating a subscriber, Err: %v", err)
	}

	log.Printf("Subscriber: %+v", subscriber)

	var wg sync.WaitGroup
	wg.Add(1)
	go subscriber.Process(ctx, &wg)
	wg.Wait()

}

// CreateSubscription creates a subscription
func CreateSubscription(ctx context.Context, subscriberConfig SubscriberConfig) (*Subscriber, error) {
	projectID := "coresystem-171219"
	client, err := getClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	topic, err := client.createTopic(ctx, subscriberConfig.TopicName)
	if err != nil {
		return nil, err
	}
	subscription, err := client.createSubscription(ctx, subscriberConfig.SubscriptionName, topic)
	if err != nil {
		return nil, err
	}
	return &Subscriber{
		topic:        topic,
		subscription: subscription,
		errorHandler: subscriberConfig.ErrorHandler,
		handle:       subscriberConfig.Handle,
	}, nil
}

func CreateJob(r request) {
	// Create Job and push the work to work queue.

	job := models.Job{
		GUID:        xid.New().String(),
		C24JobID:    r.C24JobId,
		Profile:     r.Profile,
		Action:      r.Action,
		Source:      r.Source,
		Destination: r.Destination,
		Status:      models.JobQueued, // Status queued.
	}

	if r.Action == "download" {
		// Send to work queue.
		_, err := downloadEnqueuer.Enqueue("download", work.Q{
			"guid":        job.GUID,
			"profile":     job.Profile,
			"source":      job.Source,
			"destination": job.Destination,
			"c24_job_id":  job.C24JobID,
			"action":      job.Action,
		})
		if err != nil {
			log.Fatal(err)
		}
	} else if r.Action == "transcode" {
		_, err := transcodeEnqueuer.Enqueue(r.C24JobId, work.Q{
			"guid":        job.GUID,
			"profile":     job.Profile,
			"source":      job.Source,
			"destination": job.Destination,
			"c24_job_id":  job.C24JobID,
			"action":      job.Action,
		})
		if err != nil {
			log.Fatal(err)
		}
	} else if r.Action == "snippetize" {

	} else {
		log.Fatal("Not a valid action type")
	}

	created := data.CreateJob(job)

	// Create the encode relationship.
	ed := models.EncodeData{
		JobID: created.ID,
		Progress: models.NullFloat64{
			NullFloat64: sql.NullFloat64{
				Float64: 0,
				Valid:   true,
			},
		},
		Data: models.NullString{
			NullString: sql.NullString{
				String: "{}",
				Valid:  true,
			},
		},
	}
	edCreated := data.CreateEncodeData(ed)
	created.EncodeDataID = edCreated.EncodeDataID
	actions.PrepareEncodeJob(job)
	log.Info(job)
}

// Process will start pulling from the pubsub. The process accepts a waitgroup as
// it will be easier for us to orchestrate a use case where one application needs
// more than one subscriber
func (subscriber *Subscriber) Process(ctx context.Context, wg *sync.WaitGroup) {
	log.Printf("Starting a Subscriber on topic %s", subscriber.topic.String())
	output := make(chan *pubsub.Message)
	go func(subscriber *Subscriber, output chan *pubsub.Message) {
		defer close(output)

		ctx, subscriber.cancel = context.WithCancel(ctx)
		err := subscriber.subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			output <- msg
		})

		if err != nil {
			// The wait group is stopped or marked done when an error is encountered
			subscriber.errorHandler(err)
			subscriber.stop()
			wg.Done()
		}
	}(subscriber, output)

	subscriber.handle(output)
}

// Stop the subscriber, closing the channel that was returned by Start.
func (subscriber *Subscriber) stop() {
	if subscriber.cancel != nil {
		log.Print("Stopped the subscriber")
		subscriber.cancel()
	}
}

type pubSubClient struct {
	psclient *pubsub.Client
}

func getClient(ctx context.Context, projectID string) (*pubSubClient, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("Error when creating pubsub client. Err: %v", err)
		return nil, err
	}
	return &pubSubClient{psclient: client}, nil
}

// topicExists checks if a given topic exists
func (client *pubSubClient) topicExists(ctx context.Context, topicName string) (bool, error) {
	topic := client.psclient.Topic(topicName)
	return topic.Exists(ctx)
}

// createTopic creates a topic if a topic name does not exist or returns one
// if it is already present
func (client *pubSubClient) createTopic(ctx context.Context, topicName string) (*pubsub.Topic, error) {
	topicExists, err := client.topicExists(ctx, topicName)
	if err != nil {
		log.Printf("Could not check if topic exists. Error: %+v", err)
		return nil, err
	}
	var topic *pubsub.Topic

	if !topicExists {
		topic, err = client.psclient.CreateTopic(ctx, topicName)
		if err != nil {
			log.Printf("Could not create topic. Err: %+v", err)
			return nil, err
		}
	} else {
		topic = client.psclient.Topic(topicName)
	}

	return topic, nil
}

// createSubscription creates the subscription to a topic
func (client *pubSubClient) createSubscription(ctx context.Context, subscriptionName string, topic *pubsub.Topic) (*pubsub.Subscription, error) {
	subscription := client.psclient.Subscription(subscriptionName)

	subscriptionExists, err := subscription.Exists(ctx)
	if err != nil {
		log.Printf("Could not check if subscription %s exists. Err: %v", subscriptionName, err)
		return nil, err
	}

	if !subscriptionExists {

		cfg := pubsub.SubscriptionConfig{
			Topic: topic,
			// The subscriber has a configurable, limited amount of time -- known as the ackDeadline -- to acknowledge
			// the outstanding message. Once the deadline passes, the message is no longer considered outstanding, and
			// Cloud Pub/Sub will attempt to redeliver the message.
			AckDeadline: 60 * time.Second,
		}

		subscription, err = client.psclient.CreateSubscription(ctx, subscriptionName, cfg)
		if err != nil {
			log.Printf("Could not create subscription %s. Err: %v", subscriptionName, err)
			return nil, err
		}
		subscription.ReceiveSettings = pubsub.ReceiveSettings{
			// This is the maximum amount of messages that are allowed to be processed by the callback function at a time.
			// Once this limit is reached, the client waits for messages to be acked or nacked by the callback before
			// requesting more messages from the server.
			MaxOutstandingMessages: 100,
			// This is the maximum amount of time that the client will extend a message's deadline. This value should be
			// set as high as messages are expected to be processed, plus some buffer.
			MaxExtension: 10 * time.Second,
		}
	}
	return subscription, nil
}

// subscriptionExists checks if a given subscription exists
func (client *pubSubClient) subscriptionExists(ctx context.Context, subscriptionName string) (bool, error) {
	subscription := client.psclient.Subscription(subscriptionName)
	return subscription.Exists(ctx)
}

// deleteSubscription deletes a subscription
func (client *pubSubClient) deleteSubscription(ctx context.Context, subscriptionName string) error {
	return client.psclient.Subscription(subscriptionName).Delete(ctx)
}

// listAllSubscription lists all subscriptions in the project
func (client *pubSubClient) listAllSubscription(ctx context.Context, topicName string) ([]string, error) {
	subscriptionNames := make([]string, 0)
	subscriptionIterator := client.psclient.Subscriptions(ctx)
	for {
		item, err := subscriptionIterator.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			log.Printf("Could not list all topics. Error %v", err)
			return subscriptionNames, err
		}
		subscriptionNames = append(subscriptionNames, item.String())
	}
	return subscriptionNames, nil
}
