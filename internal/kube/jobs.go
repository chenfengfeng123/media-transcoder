package kube

import (
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilpointer "k8s.io/utils/pointer"
	"math/rand"
	"sort"
	"strconv"
)

const (
	// Name of the generated sync job
	Name      = "transcode"
	imageName  = "gcr.io/coresystem-171219/c24-media:initial"
	defaultPVCName = "mpc-storage-std-claim"
	defaultCredPath = "/google-cloud.json"
	imagePullPolicy = apiv1.PullAlways
	defaultNamespace = "jobs"
	defaultJobName = "transcode"
	defaultMemoryReq = "2000M"
	defaultMemoryLim = "2000M"
	defaultCPUReq = "500m"
	defaultCPULim = "500m"
	defaultCompletions = int32(2)
	defaultParallelism = int32(2)
	defaultBackoff = int32(3)
)

// New creates a synchronize job that synchronizes secrets.
func New(options ...func(*batchv1.Job)) *batchv1.Job {
	b := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: defaultJobName,
			Namespace: defaultNamespace,
			Labels: map[string]string{
				"job": defaultJobName,
			},
		},
		Spec: batchv1.JobSpec{
			Completions: nil,
			Parallelism: nil,
			BackoffLimit: nil,
			TTLSecondsAfterFinished: nil,
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					ServiceAccountName: "default",
					RestartPolicy:      apiv1.RestartPolicyNever,
					Volumes: []apiv1.Volume{
						{
							Name: "transcode",
							VolumeSource: apiv1.VolumeSource{
								EmptyDir: &apiv1.EmptyDirVolumeSource{
									Medium: apiv1.StorageMediumMemory,
								},
							},
						},
						{
							Name:         "mpc-storage-std",
							VolumeSource: apiv1.VolumeSource{
								PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
									ClaimName: defaultPVCName,
									ReadOnly:  false,
								},
							},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:            "transcode",
							Image:			 imageName,
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "mpc-storage-std",
									MountPath: "/mpc",
								},
							},
							Resources: apiv1.ResourceRequirements{
								Limits: map[apiv1.ResourceName]resource.Quantity{
									apiv1.ResourceMemory: resource.MustParse(defaultMemoryLim),
									apiv1.ResourceCPU:    resource.MustParse(defaultCPULim),
								},
								Requests: map[apiv1.ResourceName]resource.Quantity{
									apiv1.ResourceMemory: resource.MustParse(defaultMemoryReq),
									apiv1.ResourceCPU:    resource.MustParse(defaultCPUReq),
								},
							},
							ImagePullPolicy: imagePullPolicy,
							Args:            []string{"transcoder"},
							LivenessProbe:  nil,
							ReadinessProbe: nil,
							Env: []apiv1.EnvVar{
								{
									Name:  "JOB_NAME",
									Value: defaultJobName,
								},
								{
									Name: "GOOGLE_APPLICATION_CREDENTIALS",
									Value: defaultCredPath,
								},
							},
						},
					},
				},
			},
		},
	}

	for _, opt := range options {
		opt(b)
	}
	sortEnv(b.Spec.Template.Spec.Containers[0].Env)

	return b
}

func sortEnv(env []apiv1.EnvVar) {
	sort.Slice(env, func(i, j int) bool {
		return env[i].Name < env[j].Name
	})
}

// WithImage configures the container's image.
func WithImage(image string) func(*batchv1.Job) {
	return func(b *batchv1.Job) {
		b.Spec.Template.Spec.Containers[0].Image = image
	}
}

func WithCommandArgs(action string) func(*batchv1.Job) {
	return func(b *batchv1.Job) {
		b.Spec.Template.Spec.Containers[0].Args[0] = action
	}
}

// WithImage configures the container's image.
func WithJobName(name string) func(*batchv1.Job) {
	return func(b *batchv1.Job) {
		b.GenerateName = name
	}
}

// WithImage configures the container's image.
func WithTTLCleanupTime(duration int32) func(*batchv1.Job) {
	return func(b *batchv1.Job) {
		b.Spec.TTLSecondsAfterFinished = utilpointer.Int32Ptr(duration)
	}
}

// WithJobNameEnv configures the pod to have a job name.
func WithJobNameEnv(name string) func(*batchv1.Job) {
	return func(b *batchv1.Job) {
		b.Spec.Template.Spec.Containers[0].Env[0] = apiv1.EnvVar{
			Name:      "JOB_NAME",
			Value:     name,
			ValueFrom: nil,
		}
	}
}

// WithMaxParallelism configures max executions per pod.
func WithMaxParallelism(parallelism string) func(*batchv1.Job) {
	pInt, _ := strconv.Atoi(parallelism)
	p := int32(pInt)
	return func(b *batchv1.Job) {
		b.Spec.Parallelism = utilpointer.Int32Ptr(p)
	}
}

// WithMaxRetries configures the max number of retries on failure.
func WithMaxRetries(attempts string) func(*batchv1.Job) {
	rInt, _ := strconv.Atoi(attempts)
	r := int32(rInt)
	return func(b *batchv1.Job) {
		b.Spec.BackoffLimit = utilpointer.Int32Ptr(r)
	}
}

// WithJobRetentionDuration configures how long to keep a job upon completion or failure.
func WithJobRetentionDuration(duration string) func(*batchv1.Job) {
	return func(b *batchv1.Job) {

	}
}

// WithCPURequest configures max CPU allocated to a job.
func WithResourceRequest(cpu string, memory string) func(*batchv1.Job) {
	return func(b *batchv1.Job) {
		b.Spec.Template.Spec.Containers[0].Resources.Limits = map[apiv1.ResourceName]resource.Quantity{
			apiv1.ResourceMemory: resource.MustParse(memory),
			apiv1.ResourceCPU:    resource.MustParse(cpu),
		}
	}
}

// WithCPURequest configures max memory allocated to a job.
func WithResourceLimit(cpu string, memory string) func(*batchv1.Job) {
	return func(b *batchv1.Job) {
		b.Spec.Template.Spec.Containers[0].Resources.Limits = map[apiv1.ResourceName]resource.Quantity{
			apiv1.ResourceMemory: resource.MustParse(memory),
			apiv1.ResourceCPU:    resource.MustParse(cpu),
		}
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}