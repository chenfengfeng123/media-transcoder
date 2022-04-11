package c24_media

import (
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"context"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"log"
)

var (
	OAuthConfig       *oauth2.Config
	StorageBucket     *storage.BucketHandle
	StorageBucketName string
	PubsubClient      *pubsub.Client
)

const PubsubTopicID = "c24-transcode-jobs"
const PubsubTopicSubscription = "c24-transcode-jobs-sub"

func init() {
	var err error

	// [START storage]
	// To configure Cloud Storage, uncomment the following lines and update the
	// bucket name.
	//
	StorageBucketName = "dev-experiments"
	StorageBucket, err = configureStorage(StorageBucketName)
	// [END storage]

	PubsubClient, err = configurePubsub("coresystem-171219")
	// [END pubsub]

	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func configureStorage(bucketID string) (*storage.BucketHandle, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.Bucket(bucketID), nil
}

func configurePubsub(projectID string) (*pubsub.Client, error) {

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Create the topic if it doesn't exist.
	if exists, err := client.Topic(PubsubTopicID).Exists(ctx); err != nil {
		return nil, err
	} else if !exists {
		if _, err := client.CreateTopic(ctx, PubsubTopicID); err != nil {
			return nil, err
		}
	}
	return client, nil
}

// C is a config instance available as a public config object.
var C Config

// Config defines the main configuration object.
type Config struct {
	Port                     string `mapstructure:"server_port"`
	RedisHost                string `mapstructure:"redis_host"`
	RedisPort                int    `mapstructure:"redis_port"`
	DatabaseHost             string `mapstructure:"database_host"`
	DatabasePort             int    `mapstructure:"database_port"`
	DatabaseUser             string `mapstructure:"database_user"`
	DatabasePassword         string `mapstructure:"database_password"`
	DatabaseName             string `mapstructure:"database_name"`
	DispatcherType			 string `mapstructure:"dispatcher_type"`
	DownloadWorkerNamespace  string `mapstructure:"download_worker_namespace"`
	DownloadWorkerJobName    string `mapstructure:"download_worker_job_name"`
	TranscodeWorkerNamespace string `mapstructure:"transcode_worker_namespace"`
	TranscodeWorkerJobName   string `mapstructure:"transcode_worker_job_name"`
	WorkerConcurrency        uint   `mapstructure:"worker_concurrency"`
	AWSRegion                string `mapstructure:"aws_region"`
	AWSAccessKey             string `mapstructure:"aws_access_key"`
	AWSSecretKey             string `mapstructure:"aws_secret_key"`
	S3InboundBucket          string `mapstructure:"s3_inbound_bucket"`
	S3InboundRegion          string `mapstructure:"s3_inbound_region"`
	S3OutboundBucket         string `mapstructure:"s3_outbound_bucket"`
	S3OutboundRegion         string `mapstructure:"s3_outbound_region"`
	WorkDirectory            string `mapstructure:"work_dir"`
	SlackWebhook             string `mapstructure:"slack_webhook"`
	DigitalOceanAccessToken  string `mapstructure:"digitalocean_access_token"`

	CloudinitRedisHost        string `mapstructure:"cloudinit_redis_host"`
	CloudinitRedisPort        int    `mapstructure:"cloudinit_redis_port"`
	CloudinitDatabaseHost     string `mapstructure:"cloudinit_database_host"`
	CloudinitDatabasePort     int    `mapstructure:"cloudinit_database_port"`
	CloudinitDatabaseUser     string `mapstructure:"cloudinit_database_user"`
	CloudinitDatabasePassword string `mapstructure:"cloudinit_database_password"`
	CloudinitDatabaseName     string `mapstructure:"cloudinit_database_name"`

	Profiles []profile
}

type profile struct {
	Profile string   `json:"profile"`
	Output  string   `json:"output"`
	Publish bool     `json:"publish"`
	Options []string `json:"options"`
}

// LoadConfig loads up the configuration struct.
func LoadConfig(file string) {
	viper.SetConfigType("yaml")
	viper.SetConfigName(file)
	viper.AddConfigPath(".")
	viper.AddConfigPath("config")
	err := viper.ReadInConfig()

	viper.AutomaticEnv()
	err = viper.Unmarshal(&C)
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
}

// GetFFmpegProfile finds an encoding profile by profile name.
func GetFFmpegProfile(profile string) (t *profile, err error) {
	for _, v := range C.Profiles {
		if v.Profile == profile {
			return &v, nil
		}
	}
	return nil, errors.New("No task")
}

// Get gets the current config.
func Get() *Config {
	return &C
}
