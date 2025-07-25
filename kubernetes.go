package godo

import (
	"bytes"
	"context"
	"encoding"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	kubernetesBasePath     = "/v2/kubernetes"
	kubernetesClustersPath = kubernetesBasePath + "/clusters"
	kubernetesOptionsPath  = kubernetesBasePath + "/options"
)

// KubernetesService is an interface for interfacing with the Kubernetes endpoints
// of the DigitalOcean API.
// See: https://docs.digitalocean.com/reference/api/api-reference/#tag/Kubernetes
type KubernetesService interface {
	Create(context.Context, *KubernetesClusterCreateRequest) (*KubernetesCluster, *Response, error)
	Get(context.Context, string) (*KubernetesCluster, *Response, error)
	GetUser(context.Context, string) (*KubernetesClusterUser, *Response, error)
	GetUpgrades(context.Context, string) ([]*KubernetesVersion, *Response, error)
	GetKubeConfig(context.Context, string) (*KubernetesClusterConfig, *Response, error)
	GetKubeConfigWithExpiry(context.Context, string, int64) (*KubernetesClusterConfig, *Response, error)
	GetCredentials(context.Context, string, *KubernetesClusterCredentialsGetRequest) (*KubernetesClusterCredentials, *Response, error)
	List(context.Context, *ListOptions) ([]*KubernetesCluster, *Response, error)
	Update(context.Context, string, *KubernetesClusterUpdateRequest) (*KubernetesCluster, *Response, error)
	Upgrade(context.Context, string, *KubernetesClusterUpgradeRequest) (*Response, error)
	Delete(context.Context, string) (*Response, error)
	DeleteSelective(context.Context, string, *KubernetesClusterDeleteSelectiveRequest) (*Response, error)
	DeleteDangerous(context.Context, string) (*Response, error)
	ListAssociatedResourcesForDeletion(context.Context, string) (*KubernetesAssociatedResources, *Response, error)

	CreateNodePool(ctx context.Context, clusterID string, req *KubernetesNodePoolCreateRequest) (*KubernetesNodePool, *Response, error)
	GetNodePool(ctx context.Context, clusterID, poolID string) (*KubernetesNodePool, *Response, error)
	GetNodePoolTemplate(ctx context.Context, clusterID string, nodePoolName string) (*KubernetesNodePoolTemplate, *Response, error)
	ListNodePools(ctx context.Context, clusterID string, opts *ListOptions) ([]*KubernetesNodePool, *Response, error)
	UpdateNodePool(ctx context.Context, clusterID, poolID string, req *KubernetesNodePoolUpdateRequest) (*KubernetesNodePool, *Response, error)
	// RecycleNodePoolNodes is DEPRECATED please use DeleteNode
	// The method will be removed in godo 2.0.
	RecycleNodePoolNodes(ctx context.Context, clusterID, poolID string, req *KubernetesNodePoolRecycleNodesRequest) (*Response, error)
	DeleteNodePool(ctx context.Context, clusterID, poolID string) (*Response, error)
	DeleteNode(ctx context.Context, clusterID, poolID, nodeID string, req *KubernetesNodeDeleteRequest) (*Response, error)

	GetOptions(context.Context) (*KubernetesOptions, *Response, error)
	AddRegistry(ctx context.Context, req *KubernetesClusterRegistryRequest) (*Response, error)
	RemoveRegistry(ctx context.Context, req *KubernetesClusterRegistryRequest) (*Response, error)

	RunClusterlint(ctx context.Context, clusterID string, req *KubernetesRunClusterlintRequest) (string, *Response, error)
	GetClusterlintResults(ctx context.Context, clusterID string, req *KubernetesGetClusterlintRequest) ([]*ClusterlintDiagnostic, *Response, error)

	GetClusterStatusMessages(ctx context.Context, clusterID string, req *KubernetesGetClusterStatusMessagesRequest) ([]*KubernetesClusterStatusMessage, *Response, error)
}

var _ KubernetesService = &KubernetesServiceOp{}

// KubernetesServiceOp handles communication with Kubernetes methods of the DigitalOcean API.
type KubernetesServiceOp struct {
	client *Client
}

// KubernetesClusterCreateRequest represents a request to create a Kubernetes cluster.
type KubernetesClusterCreateRequest struct {
	Name          string   `json:"name,omitempty"`
	RegionSlug    string   `json:"region,omitempty"`
	VersionSlug   string   `json:"version,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	VPCUUID       string   `json:"vpc_uuid,omitempty"`
	ClusterSubnet string   `json:"cluster_subnet,omitempty"`
	ServiceSubnet string   `json:"service_subnet,omitempty"`

	// Create cluster with highly available control plane
	HA bool `json:"ha"`

	NodePools []*KubernetesNodePoolCreateRequest `json:"node_pools,omitempty"`

	MaintenancePolicy                 *KubernetesMaintenancePolicy                 `json:"maintenance_policy"`
	AutoUpgrade                       bool                                         `json:"auto_upgrade"`
	SurgeUpgrade                      bool                                         `json:"surge_upgrade"`
	ControlPlaneFirewall              *KubernetesControlPlaneFirewall              `json:"control_plane_firewall,omitempty"`
	ClusterAutoscalerConfiguration    *KubernetesClusterAutoscalerConfiguration    `json:"cluster_autoscaler_configuration,omitempty"`
	RoutingAgent                      *KubernetesRoutingAgent                      `json:"routing_agent,omitempty"`
	AmdGpuDevicePlugin                *KubernetesAmdGpuDevicePlugin                `json:"amd_gpu_device_plugin,omitempty"`
	AmdGpuDeviceMetricsExporterPlugin *KubernetesAmdGpuDeviceMetricsExporterPlugin `json:"amd_gpu_device_metrics_exporter_plugin,omitempty"`
}

// KubernetesClusterUpdateRequest represents a request to update a Kubernetes cluster.
type KubernetesClusterUpdateRequest struct {
	Name                              string                                       `json:"name,omitempty"`
	Tags                              []string                                     `json:"tags,omitempty"`
	MaintenancePolicy                 *KubernetesMaintenancePolicy                 `json:"maintenance_policy,omitempty"`
	AutoUpgrade                       *bool                                        `json:"auto_upgrade,omitempty"`
	SurgeUpgrade                      bool                                         `json:"surge_upgrade,omitempty"`
	ControlPlaneFirewall              *KubernetesControlPlaneFirewall              `json:"control_plane_firewall,omitempty"`
	ClusterAutoscalerConfiguration    *KubernetesClusterAutoscalerConfiguration    `json:"cluster_autoscaler_configuration,omitempty"`
	RoutingAgent                      *KubernetesRoutingAgent                      `json:"routing_agent,omitempty"`
	AmdGpuDevicePlugin                *KubernetesAmdGpuDevicePlugin                `json:"amd_gpu_device_plugin,omitempty"`
	AmdGpuDeviceMetricsExporterPlugin *KubernetesAmdGpuDeviceMetricsExporterPlugin `json:"amd_gpu_device_metrics_exporter_plugin,omitempty"`

	// Convert cluster to run highly available control plane
	HA *bool `json:"ha,omitempty"`
}

// KubernetesClusterDeleteSelectiveRequest represents a delete selective request to delete a cluster and it's associated resources.
type KubernetesClusterDeleteSelectiveRequest struct {
	Volumes         []string `json:"volumes"`
	VolumeSnapshots []string `json:"volume_snapshots"`
	LoadBalancers   []string `json:"load_balancers"`
}

// KubernetesClusterUpgradeRequest represents a request to upgrade a Kubernetes cluster.
type KubernetesClusterUpgradeRequest struct {
	VersionSlug string `json:"version,omitempty"`
}

// Taint represents a Kubernetes taint that can be associated with a node pool
// (and, transitively, with all nodes of that pool).
type Taint struct {
	Key    string
	Value  string
	Effect string
}

func (t Taint) String() string {
	if t.Value == "" {
		return fmt.Sprintf("%s:%s", t.Key, t.Effect)
	}
	return fmt.Sprintf("%s=%s:%s", t.Key, t.Value, t.Effect)
}

// KubernetesNodePoolCreateRequest represents a request to create a node pool for a
// Kubernetes cluster.
type KubernetesNodePoolCreateRequest struct {
	Name      string            `json:"name,omitempty"`
	Size      string            `json:"size,omitempty"`
	Count     int               `json:"count,omitempty"`
	Tags      []string          `json:"tags,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	Taints    []Taint           `json:"taints,omitempty"`
	AutoScale bool              `json:"auto_scale,omitempty"`
	MinNodes  int               `json:"min_nodes,omitempty"`
	MaxNodes  int               `json:"max_nodes,omitempty"`
}

// KubernetesNodePoolUpdateRequest represents a request to update a node pool in a
// Kubernetes cluster.
type KubernetesNodePoolUpdateRequest struct {
	Name      string            `json:"name,omitempty"`
	Count     *int              `json:"count,omitempty"`
	Tags      []string          `json:"tags,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	Taints    *[]Taint          `json:"taints,omitempty"`
	AutoScale *bool             `json:"auto_scale,omitempty"`
	MinNodes  *int              `json:"min_nodes,omitempty"`
	MaxNodes  *int              `json:"max_nodes,omitempty"`
}

// KubernetesNodePoolRecycleNodesRequest is DEPRECATED please use DeleteNode
// The type will be removed in godo 2.0.
type KubernetesNodePoolRecycleNodesRequest struct {
	Nodes []string `json:"nodes,omitempty"`
}

// KubernetesNodeDeleteRequest is a request to delete a specific node in a node pool.
type KubernetesNodeDeleteRequest struct {
	// Replace will cause a new node to be created to replace the deleted node.
	Replace bool `json:"replace,omitempty"`

	// SkipDrain skips draining the node before deleting it.
	SkipDrain bool `json:"skip_drain,omitempty"`
}

// KubernetesClusterCredentialsGetRequest is a request to get cluster credentials.
type KubernetesClusterCredentialsGetRequest struct {
	ExpirySeconds *int `json:"expiry_seconds,omitempty"`
}

// KubernetesClusterRegistryRequest represents clusters to integrate with docr registry
type KubernetesClusterRegistryRequest struct {
	ClusterUUIDs []string `json:"cluster_uuids,omitempty"`
}

type KubernetesRunClusterlintRequest struct {
	IncludeGroups []string `json:"include_groups"`
	ExcludeGroups []string `json:"exclude_groups"`
	IncludeChecks []string `json:"include_checks"`
	ExcludeChecks []string `json:"exclude_checks"`
}

type KubernetesGetClusterlintRequest struct {
	RunId string `json:"run_id"`
}

type clusterStatusMessagesRoot struct {
	Messages []*KubernetesClusterStatusMessage `json:"messages"`
}

type KubernetesClusterStatusMessage struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type KubernetesGetClusterStatusMessagesRequest struct {
	Since *time.Time `json:"since"`
}

// KubernetesCluster represents a Kubernetes cluster.
type KubernetesCluster struct {
	ID            string   `json:"id,omitempty"`
	Name          string   `json:"name,omitempty"`
	RegionSlug    string   `json:"region,omitempty"`
	VersionSlug   string   `json:"version,omitempty"`
	ClusterSubnet string   `json:"cluster_subnet,omitempty"`
	ServiceSubnet string   `json:"service_subnet,omitempty"`
	IPv4          string   `json:"ipv4,omitempty"`
	Endpoint      string   `json:"endpoint,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	VPCUUID       string   `json:"vpc_uuid,omitempty"`

	// Cluster runs a highly available control plane
	HA bool `json:"ha,omitempty"`

	NodePools []*KubernetesNodePool `json:"node_pools,omitempty"`

	MaintenancePolicy                 *KubernetesMaintenancePolicy                 `json:"maintenance_policy,omitempty"`
	AutoUpgrade                       bool                                         `json:"auto_upgrade,omitempty"`
	SurgeUpgrade                      bool                                         `json:"surge_upgrade,omitempty"`
	RegistryEnabled                   bool                                         `json:"registry_enabled,omitempty"`
	ControlPlaneFirewall              *KubernetesControlPlaneFirewall              `json:"control_plane_firewall,omitempty"`
	ClusterAutoscalerConfiguration    *KubernetesClusterAutoscalerConfiguration    `json:"cluster_autoscaler_configuration,omitempty"`
	RoutingAgent                      *KubernetesRoutingAgent                      `json:"routing_agent,omitempty"`
	AmdGpuDevicePlugin                *KubernetesAmdGpuDevicePlugin                `json:"amd_gpu_device_plugin,omitempty"`
	AmdGpuDeviceMetricsExporterPlugin *KubernetesAmdGpuDeviceMetricsExporterPlugin `json:"amd_gpu_device_metrics_exporter_plugin,omitempty"`

	Status    *KubernetesClusterStatus `json:"status,omitempty"`
	CreatedAt time.Time                `json:"created_at,omitempty"`
	UpdatedAt time.Time                `json:"updated_at,omitempty"`
}

// URN returns the Kubernetes cluster's ID in the format of DigitalOcean URN.
func (kc KubernetesCluster) URN() string {
	return ToURN("Kubernetes", kc.ID)
}

// KubernetesClusterUser represents a Kubernetes cluster user.
type KubernetesClusterUser struct {
	ID       string   `json:"id,omitempty"`
	Username string   `json:"username,omitempty"`
	Groups   []string `json:"groups,omitempty"`
}

// KubernetesClusterCredentials represents Kubernetes cluster credentials.
type KubernetesClusterCredentials struct {
	Server                   string    `json:"server"`
	CertificateAuthorityData []byte    `json:"certificate_authority_data"`
	ClientCertificateData    []byte    `json:"client_certificate_data"`
	ClientKeyData            []byte    `json:"client_key_data"`
	Token                    string    `json:"token"`
	ExpiresAt                time.Time `json:"expires_at"`
}

// KubernetesMaintenancePolicy is a configuration to set the maintenance window
// of a cluster
type KubernetesMaintenancePolicy struct {
	StartTime string                         `json:"start_time"`
	Duration  string                         `json:"duration"`
	Day       KubernetesMaintenancePolicyDay `json:"day"`
}

// KubernetesControlPlaneFirewall represents Kubernetes cluster control plane firewall.
type KubernetesControlPlaneFirewall struct {
	Enabled          *bool    `json:"enabled"`
	AllowedAddresses []string `json:"allowed_addresses"`
}

// KubernetesRoutingAgent represents information about the routing-agent cluster plugin.
type KubernetesRoutingAgent struct {
	Enabled *bool `json:"enabled"`
}

// KubernetesAmdGpuDevicePlugin represents information about the AMD GPU Device Plugin cluster plugin.
// If a cluster has a node pool with an AMD GPU it will be enabled by default.
type KubernetesAmdGpuDevicePlugin struct {
	Enabled *bool `json:"enabled"`
}

// KubernetesAmdGpuDeviceMetricsExporterPlugin represents information about the AMD GPU Device Metrics Exporter cluster plugin.
type KubernetesAmdGpuDeviceMetricsExporterPlugin struct {
	Enabled *bool `json:"enabled"`
}

// KubernetesClusterAutoscalerConfiguration represents Kubernetes cluster autoscaler configuration.
type KubernetesClusterAutoscalerConfiguration struct {
	ScaleDownUtilizationThreshold *float64 `json:"scale_down_utilization_threshold"`
	ScaleDownUnneededTime         *string  `json:"scale_down_unneeded_time"`
	Expanders                     []string `json:"expanders"`
}

// KubernetesMaintenancePolicyDay represents the possible days of a maintenance
// window
type KubernetesMaintenancePolicyDay int

const (
	// KubernetesMaintenanceDayAny sets the KubernetesMaintenancePolicyDay to any
	// day of the week
	KubernetesMaintenanceDayAny KubernetesMaintenancePolicyDay = iota

	// KubernetesMaintenanceDayMonday sets the KubernetesMaintenancePolicyDay to
	// Monday
	KubernetesMaintenanceDayMonday

	// KubernetesMaintenanceDayTuesday sets the KubernetesMaintenancePolicyDay to
	// Tuesday
	KubernetesMaintenanceDayTuesday

	// KubernetesMaintenanceDayWednesday sets the KubernetesMaintenancePolicyDay to
	// Wednesday
	KubernetesMaintenanceDayWednesday

	// KubernetesMaintenanceDayThursday sets the KubernetesMaintenancePolicyDay to
	// Thursday
	KubernetesMaintenanceDayThursday

	// KubernetesMaintenanceDayFriday sets the KubernetesMaintenancePolicyDay to
	// Friday
	KubernetesMaintenanceDayFriday

	// KubernetesMaintenanceDaySaturday sets the KubernetesMaintenancePolicyDay to
	// Saturday
	KubernetesMaintenanceDaySaturday

	// KubernetesMaintenanceDaySunday sets the KubernetesMaintenancePolicyDay to
	// Sunday
	KubernetesMaintenanceDaySunday
)

var (
	days = [...]string{
		"any",
		"monday",
		"tuesday",
		"wednesday",
		"thursday",
		"friday",
		"saturday",
		"sunday",
	}

	toDay = map[string]KubernetesMaintenancePolicyDay{
		"any":       KubernetesMaintenanceDayAny,
		"monday":    KubernetesMaintenanceDayMonday,
		"tuesday":   KubernetesMaintenanceDayTuesday,
		"wednesday": KubernetesMaintenanceDayWednesday,
		"thursday":  KubernetesMaintenanceDayThursday,
		"friday":    KubernetesMaintenanceDayFriday,
		"saturday":  KubernetesMaintenanceDaySaturday,
		"sunday":    KubernetesMaintenanceDaySunday,
	}
)

// KubernetesMaintenanceToDay returns the appropriate KubernetesMaintenancePolicyDay for the given string.
func KubernetesMaintenanceToDay(day string) (KubernetesMaintenancePolicyDay, error) {
	d, ok := toDay[strings.ToLower(day)]
	if !ok {
		return 0, fmt.Errorf("unknown day: %q", day)
	}

	return d, nil
}

func (k KubernetesMaintenancePolicyDay) String() string {
	if KubernetesMaintenanceDayAny <= k && k <= KubernetesMaintenanceDaySunday {
		return days[k]
	}
	return fmt.Sprintf("%d !Weekday", k)

}

// UnmarshalJSON parses the JSON string into KubernetesMaintenancePolicyDay
func (k *KubernetesMaintenancePolicyDay) UnmarshalJSON(data []byte) error {
	var val string
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}

	parsed, err := KubernetesMaintenanceToDay(val)
	if err != nil {
		return err
	}
	*k = parsed
	return nil
}

// MarshalJSON returns the JSON string for KubernetesMaintenancePolicyDay
func (k KubernetesMaintenancePolicyDay) MarshalJSON() ([]byte, error) {
	if KubernetesMaintenanceDayAny <= k && k <= KubernetesMaintenanceDaySunday {
		return json.Marshal(days[k])
	}

	return nil, fmt.Errorf("invalid day: %d", k)
}

// Possible states for a cluster.
const (
	KubernetesClusterStatusProvisioning = KubernetesClusterStatusState("provisioning")
	KubernetesClusterStatusRunning      = KubernetesClusterStatusState("running")
	KubernetesClusterStatusDegraded     = KubernetesClusterStatusState("degraded")
	KubernetesClusterStatusError        = KubernetesClusterStatusState("error")
	KubernetesClusterStatusDeleted      = KubernetesClusterStatusState("deleted")
	KubernetesClusterStatusUpgrading    = KubernetesClusterStatusState("upgrading")
	KubernetesClusterStatusInvalid      = KubernetesClusterStatusState("invalid")
)

// KubernetesClusterStatusState represents states for a cluster.
type KubernetesClusterStatusState string

var _ encoding.TextUnmarshaler = (*KubernetesClusterStatusState)(nil)

// UnmarshalText unmarshals the state.
func (s *KubernetesClusterStatusState) UnmarshalText(text []byte) error {
	switch KubernetesClusterStatusState(strings.ToLower(string(text))) {
	case KubernetesClusterStatusProvisioning:
		*s = KubernetesClusterStatusProvisioning
	case KubernetesClusterStatusRunning:
		*s = KubernetesClusterStatusRunning
	case KubernetesClusterStatusDegraded:
		*s = KubernetesClusterStatusDegraded
	case KubernetesClusterStatusError:
		*s = KubernetesClusterStatusError
	case KubernetesClusterStatusDeleted:
		*s = KubernetesClusterStatusDeleted
	case KubernetesClusterStatusUpgrading:
		*s = KubernetesClusterStatusUpgrading
	case "", KubernetesClusterStatusInvalid:
		*s = KubernetesClusterStatusInvalid
	default:
		return fmt.Errorf("unknown cluster state %q", string(text))
	}
	return nil
}

// KubernetesClusterStatus describes the status of a cluster.
type KubernetesClusterStatus struct {
	State   KubernetesClusterStatusState `json:"state,omitempty"`
	Message string                       `json:"message,omitempty"`
}

// KubernetesNodePool represents a node pool in a Kubernetes cluster.
type KubernetesNodePool struct {
	ID        string            `json:"id,omitempty"`
	Name      string            `json:"name,omitempty"`
	Size      string            `json:"size,omitempty"`
	Count     int               `json:"count,omitempty"`
	Tags      []string          `json:"tags,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	Taints    []Taint           `json:"taints,omitempty"`
	AutoScale bool              `json:"auto_scale,omitempty"`
	MinNodes  int               `json:"min_nodes,omitempty"`
	MaxNodes  int               `json:"max_nodes,omitempty"`

	Nodes []*KubernetesNode `json:"nodes,omitempty"`
}

// KubernetesNodePool represents the node pool template data for a given pool.
type KubernetesNodePoolTemplate struct {
	Template *KubernetesNodeTemplate
}

// KubernetesNodePoolResources represents the resources within a given template for a node pool
// This follows https://pkg.go.dev/k8s.io/kubernetes@v1.32.1/pkg/scheduler/framework#Resource to represent
// node resources within the node object.
type KubernetesNodePoolResources struct {
	CPU    int64  `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
	Pods   int64  `json:"pods,omitempty"`
}

// KubernetesNode represents a Node in a node pool in a Kubernetes cluster.
type KubernetesNode struct {
	ID        string                `json:"id,omitempty"`
	Name      string                `json:"name,omitempty"`
	Status    *KubernetesNodeStatus `json:"status,omitempty"`
	DropletID string                `json:"droplet_id,omitempty"`

	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// KubernetesNodeTemplate represents a template in a node pool in a Kubernetes cluster.
type KubernetesNodeTemplate struct {
	ClusterUUID string                       `json:"cluster_uuid,omitempty"`
	Name        string                       `json:"name,omitempty"`
	Slug        string                       `json:"slug,omitempty"`
	Labels      map[string]string            `json:"labels,omitempty"`
	Taints      []string                     `json:"taints,omitempty"`
	Capacity    *KubernetesNodePoolResources `json:"capacity,omitempty"`
	Allocatable *KubernetesNodePoolResources `json:"allocatable,omitempty"`
}

// KubernetesNodeStatus represents the status of a particular Node in a Kubernetes cluster.
type KubernetesNodeStatus struct {
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}

// KubernetesOptions represents options available for creating Kubernetes clusters.
type KubernetesOptions struct {
	Versions []*KubernetesVersion  `json:"versions,omitempty"`
	Regions  []*KubernetesRegion   `json:"regions,omitempty"`
	Sizes    []*KubernetesNodeSize `json:"sizes,omitempty"`
}

// KubernetesVersion is a DigitalOcean Kubernetes release.
type KubernetesVersion struct {
	Slug              string   `json:"slug,omitempty"`
	KubernetesVersion string   `json:"kubernetes_version,omitempty"`
	SupportedFeatures []string `json:"supported_features,omitempty"`
}

// KubernetesNodeSize is a node sizes supported for Kubernetes clusters.
type KubernetesNodeSize struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// KubernetesRegion is a region usable by Kubernetes clusters.
type KubernetesRegion struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// ClusterlintDiagnostic is a diagnostic returned from clusterlint.
type ClusterlintDiagnostic struct {
	CheckName string             `json:"check_name"`
	Severity  string             `json:"severity"`
	Message   string             `json:"message"`
	Object    *ClusterlintObject `json:"object"`
}

// ClusterlintObject is the object a clusterlint diagnostic refers to.
type ClusterlintObject struct {
	Kind      string              `json:"kind"`
	Name      string              `json:"name"`
	Namespace string              `json:"namespace"`
	Owners    []*ClusterlintOwner `json:"owners,omitempty"`
}

// ClusterlintOwner indicates the resource that owns the offending object.
type ClusterlintOwner struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

// KubernetesAssociatedResources represents a cluster's associated resources
type KubernetesAssociatedResources struct {
	Volumes         []*AssociatedResource `json:"volumes"`
	VolumeSnapshots []*AssociatedResource `json:"volume_snapshots"`
	LoadBalancers   []*AssociatedResource `json:"load_balancers"`
}

// AssociatedResource is the object to represent a Kubernetes cluster associated resource's ID and Name.
type AssociatedResource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type kubernetesClustersRoot struct {
	Clusters []*KubernetesCluster `json:"kubernetes_clusters,omitempty"`
	Links    *Links               `json:"links,omitempty"`
	Meta     *Meta                `json:"meta"`
}

type kubernetesClusterRoot struct {
	Cluster *KubernetesCluster `json:"kubernetes_cluster,omitempty"`
}

type kubernetesClusterUserRoot struct {
	User *KubernetesClusterUser `json:"kubernetes_cluster_user,omitempty"`
}

type kubernetesNodePoolRoot struct {
	NodePool *KubernetesNodePool `json:"node_pool,omitempty"`
}

type kubernetesNodePoolsRoot struct {
	NodePools []*KubernetesNodePool `json:"node_pools,omitempty"`
	Links     *Links                `json:"links,omitempty"`
}

type kubernetesUpgradesRoot struct {
	AvailableUpgradeVersions []*KubernetesVersion `json:"available_upgrade_versions,omitempty"`
}

// Get retrieves the details of a Kubernetes cluster.
func (svc *KubernetesServiceOp) Get(ctx context.Context, clusterID string) (*KubernetesCluster, *Response, error) {
	path := fmt.Sprintf("%s/%s", kubernetesClustersPath, clusterID)
	req, err := svc.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(kubernetesClusterRoot)
	resp, err := svc.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.Cluster, resp, nil
}

// GetUser retrieves the details of a Kubernetes cluster user.
func (svc *KubernetesServiceOp) GetUser(ctx context.Context, clusterID string) (*KubernetesClusterUser, *Response, error) {
	path := fmt.Sprintf("%s/%s/user", kubernetesClustersPath, clusterID)
	req, err := svc.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(kubernetesClusterUserRoot)
	resp, err := svc.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.User, resp, nil
}

// GetUpgrades retrieves versions a Kubernetes cluster can be upgraded to. An
// upgrade can be requested using `Upgrade`.
func (svc *KubernetesServiceOp) GetUpgrades(ctx context.Context, clusterID string) ([]*KubernetesVersion, *Response, error) {
	path := fmt.Sprintf("%s/%s/upgrades", kubernetesClustersPath, clusterID)
	req, err := svc.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(kubernetesUpgradesRoot)
	resp, err := svc.client.Do(ctx, req, root)
	if err != nil {
		return nil, nil, err
	}
	return root.AvailableUpgradeVersions, resp, nil
}

// Create creates a Kubernetes cluster.
func (svc *KubernetesServiceOp) Create(ctx context.Context, create *KubernetesClusterCreateRequest) (*KubernetesCluster, *Response, error) {
	path := kubernetesClustersPath
	req, err := svc.client.NewRequest(ctx, http.MethodPost, path, create)
	if err != nil {
		return nil, nil, err
	}
	root := new(kubernetesClusterRoot)
	resp, err := svc.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.Cluster, resp, nil
}

// Delete deletes a Kubernetes cluster. There is no way to recover a cluster
// once it has been destroyed.
func (svc *KubernetesServiceOp) Delete(ctx context.Context, clusterID string) (*Response, error) {
	path := fmt.Sprintf("%s/%s", kubernetesClustersPath, clusterID)
	req, err := svc.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := svc.client.Do(ctx, req, nil)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// DeleteSelective deletes a Kubernetes cluster and the specified associated resources.
// Users can choose to delete specific volumes, volume snapshots or load balancers along with the cluster
// There is no way to recover a cluster or the specified resources once destroyed.
func (svc *KubernetesServiceOp) DeleteSelective(ctx context.Context, clusterID string, request *KubernetesClusterDeleteSelectiveRequest) (*Response, error) {
	path := fmt.Sprintf("%s/%s/destroy_with_associated_resources/selective", kubernetesClustersPath, clusterID)
	req, err := svc.client.NewRequest(ctx, http.MethodDelete, path, request)
	if err != nil {
		return nil, err
	}
	resp, err := svc.client.Do(ctx, req, nil)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// DeleteDangerous deletes a Kubernetes cluster and all its associated resources. There is no way to recover a cluster
// or it's associated resources once destroyed.
func (svc *KubernetesServiceOp) DeleteDangerous(ctx context.Context, clusterID string) (*Response, error) {
	path := fmt.Sprintf("%s/%s/destroy_with_associated_resources/dangerous", kubernetesClustersPath, clusterID)
	req, err := svc.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := svc.client.Do(ctx, req, nil)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// ListAssociatedResourcesForDeletion lists a Kubernetes cluster's resources that can be selected
// for deletion along with the cluster. See DeleteSelective
// Associated resources include volumes, volume snapshots and load balancers.
func (svc *KubernetesServiceOp) ListAssociatedResourcesForDeletion(ctx context.Context, clusterID string) (*KubernetesAssociatedResources, *Response, error) {
	path := fmt.Sprintf("%s/%s/destroy_with_associated_resources", kubernetesClustersPath, clusterID)
	req, err := svc.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(KubernetesAssociatedResources)
	resp, err := svc.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root, resp, nil
}

// List returns a list of the Kubernetes clusters visible with the caller's API token.
func (svc *KubernetesServiceOp) List(ctx context.Context, opts *ListOptions) ([]*KubernetesCluster, *Response, error) {
	path := kubernetesClustersPath
	path, err := addOptions(path, opts)
	if err != nil {
		return nil, nil, err
	}
	req, err := svc.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(kubernetesClustersRoot)
	resp, err := svc.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	if l := root.Links; l != nil {
		resp.Links = l
	}
	if m := root.Meta; m != nil {
		resp.Meta = m
	}

	return root.Clusters, resp, nil
}

// KubernetesClusterConfig is the content of a Kubernetes config file, which can be
// used to interact with your Kubernetes cluster using `kubectl`.
// See: https://kubernetes.io/docs/tasks/tools/install-kubectl/
type KubernetesClusterConfig struct {
	KubeconfigYAML []byte
}

// GetKubeConfig returns a Kubernetes config file for the specified cluster.
func (svc *KubernetesServiceOp) GetKubeConfig(ctx context.Context, clusterID string) (*KubernetesClusterConfig, *Response, error) {
	path := fmt.Sprintf("%s/%s/kubeconfig", kubernetesClustersPath, clusterID)
	req, err := svc.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	configBytes := bytes.NewBuffer(nil)
	resp, err := svc.client.Do(ctx, req, configBytes)
	if err != nil {
		return nil, resp, err
	}
	res := &KubernetesClusterConfig{
		KubeconfigYAML: configBytes.Bytes(),
	}
	return res, resp, nil
}

// GetKubeConfigWithExpiry returns a Kubernetes config file for the specified cluster with expiry_seconds.
func (svc *KubernetesServiceOp) GetKubeConfigWithExpiry(ctx context.Context, clusterID string, expirySeconds int64) (*KubernetesClusterConfig, *Response, error) {
	path := fmt.Sprintf("%s/%s/kubeconfig", kubernetesClustersPath, clusterID)
	req, err := svc.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	q := req.URL.Query()
	q.Add("expiry_seconds", fmt.Sprintf("%d", expirySeconds))
	req.URL.RawQuery = q.Encode()
	configBytes := bytes.NewBuffer(nil)
	resp, err := svc.client.Do(ctx, req, configBytes)
	if err != nil {
		return nil, resp, err
	}
	res := &KubernetesClusterConfig{
		KubeconfigYAML: configBytes.Bytes(),
	}
	return res, resp, nil
}

// GetCredentials returns a Kubernetes API server credentials for the specified cluster.
func (svc *KubernetesServiceOp) GetCredentials(ctx context.Context, clusterID string, get *KubernetesClusterCredentialsGetRequest) (*KubernetesClusterCredentials, *Response, error) {
	path := fmt.Sprintf("%s/%s/credentials", kubernetesClustersPath, clusterID)
	req, err := svc.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	q := req.URL.Query()
	if get.ExpirySeconds != nil {
		q.Add("expiry_seconds", strconv.Itoa(*get.ExpirySeconds))
	}
	req.URL.RawQuery = q.Encode()
	credentials := new(KubernetesClusterCredentials)
	resp, err := svc.client.Do(ctx, req, credentials)
	if err != nil {
		return nil, nil, err
	}
	return credentials, resp, nil
}

// Update updates a Kubernetes cluster's properties.
func (svc *KubernetesServiceOp) Update(ctx context.Context, clusterID string, update *KubernetesClusterUpdateRequest) (*KubernetesCluster, *Response, error) {
	path := fmt.Sprintf("%s/%s", kubernetesClustersPath, clusterID)
	req, err := svc.client.NewRequest(ctx, http.MethodPut, path, update)
	if err != nil {
		return nil, nil, err
	}
	root := new(kubernetesClusterRoot)
	resp, err := svc.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.Cluster, resp, nil
}

// Upgrade upgrades a Kubernetes cluster to a new version. Valid upgrade
// versions for a given cluster can be retrieved with `GetUpgrades`.
func (svc *KubernetesServiceOp) Upgrade(ctx context.Context, clusterID string, upgrade *KubernetesClusterUpgradeRequest) (*Response, error) {
	path := fmt.Sprintf("%s/%s/upgrade", kubernetesClustersPath, clusterID)
	req, err := svc.client.NewRequest(ctx, http.MethodPost, path, upgrade)
	if err != nil {
		return nil, err
	}
	return svc.client.Do(ctx, req, nil)
}

// CreateNodePool creates a new node pool in an existing Kubernetes cluster.
func (svc *KubernetesServiceOp) CreateNodePool(ctx context.Context, clusterID string, create *KubernetesNodePoolCreateRequest) (*KubernetesNodePool, *Response, error) {
	path := fmt.Sprintf("%s/%s/node_pools", kubernetesClustersPath, clusterID)
	req, err := svc.client.NewRequest(ctx, http.MethodPost, path, create)
	if err != nil {
		return nil, nil, err
	}
	root := new(kubernetesNodePoolRoot)
	resp, err := svc.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.NodePool, resp, nil
}

// GetNodePool retrieves an existing node pool in a Kubernetes cluster.
func (svc *KubernetesServiceOp) GetNodePool(ctx context.Context, clusterID, poolID string) (*KubernetesNodePool, *Response, error) {
	path := fmt.Sprintf("%s/%s/node_pools/%s", kubernetesClustersPath, clusterID, poolID)
	req, err := svc.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(kubernetesNodePoolRoot)
	resp, err := svc.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.NodePool, resp, nil
}

// GetNodePoolTemplate retrieves the template used for a given node pool to scale up from zero.
func (svc *KubernetesServiceOp) GetNodePoolTemplate(ctx context.Context, clusterID string, nodePoolName string) (*KubernetesNodePoolTemplate, *Response, error) {
	path, err := url.JoinPath(kubernetesClustersPath, clusterID, "node_pools_template", nodePoolName)
	if err != nil {
		return nil, nil, err
	}
	req, err := svc.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(KubernetesNodePoolTemplate)
	resp, err := svc.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root, resp, nil
}

// ListNodePools lists all the node pools found in a Kubernetes cluster.
func (svc *KubernetesServiceOp) ListNodePools(ctx context.Context, clusterID string, opts *ListOptions) ([]*KubernetesNodePool, *Response, error) {
	path := fmt.Sprintf("%s/%s/node_pools", kubernetesClustersPath, clusterID)
	path, err := addOptions(path, opts)
	if err != nil {
		return nil, nil, err
	}
	req, err := svc.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(kubernetesNodePoolsRoot)
	resp, err := svc.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.NodePools, resp, nil
}

// UpdateNodePool updates the details of an existing node pool.
func (svc *KubernetesServiceOp) UpdateNodePool(ctx context.Context, clusterID, poolID string, update *KubernetesNodePoolUpdateRequest) (*KubernetesNodePool, *Response, error) {
	path := fmt.Sprintf("%s/%s/node_pools/%s", kubernetesClustersPath, clusterID, poolID)
	req, err := svc.client.NewRequest(ctx, http.MethodPut, path, update)
	if err != nil {
		return nil, nil, err
	}
	root := new(kubernetesNodePoolRoot)
	resp, err := svc.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.NodePool, resp, nil
}

// RecycleNodePoolNodes is DEPRECATED please use DeleteNode
// The method will be removed in godo 2.0.
func (svc *KubernetesServiceOp) RecycleNodePoolNodes(ctx context.Context, clusterID, poolID string, recycle *KubernetesNodePoolRecycleNodesRequest) (*Response, error) {
	path := fmt.Sprintf("%s/%s/node_pools/%s/recycle", kubernetesClustersPath, clusterID, poolID)
	req, err := svc.client.NewRequest(ctx, http.MethodPost, path, recycle)
	if err != nil {
		return nil, err
	}
	resp, err := svc.client.Do(ctx, req, nil)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// DeleteNodePool deletes a node pool, and subsequently all the nodes in that pool.
func (svc *KubernetesServiceOp) DeleteNodePool(ctx context.Context, clusterID, poolID string) (*Response, error) {
	path := fmt.Sprintf("%s/%s/node_pools/%s", kubernetesClustersPath, clusterID, poolID)
	req, err := svc.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := svc.client.Do(ctx, req, nil)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// DeleteNode deletes a specific node in a node pool.
func (svc *KubernetesServiceOp) DeleteNode(ctx context.Context, clusterID, poolID, nodeID string, deleteReq *KubernetesNodeDeleteRequest) (*Response, error) {
	path := fmt.Sprintf("%s/%s/node_pools/%s/nodes/%s", kubernetesClustersPath, clusterID, poolID, nodeID)
	if deleteReq != nil {
		v := make(url.Values)
		if deleteReq.SkipDrain {
			v.Set("skip_drain", "1")
		}
		if deleteReq.Replace {
			v.Set("replace", "1")
		}
		if query := v.Encode(); query != "" {
			path = path + "?" + query
		}
	}

	req, err := svc.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := svc.client.Do(ctx, req, nil)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

type kubernetesOptionsRoot struct {
	Options *KubernetesOptions `json:"options,omitempty"`
	Links   *Links             `json:"links,omitempty"`
}

// GetOptions returns options about the Kubernetes service, such as the versions available for
// cluster creation.
func (svc *KubernetesServiceOp) GetOptions(ctx context.Context) (*KubernetesOptions, *Response, error) {
	path := kubernetesOptionsPath
	req, err := svc.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(kubernetesOptionsRoot)
	resp, err := svc.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.Options, resp, nil
}

// AddRegistry integrates docr registry with all the specified clusters
func (svc *KubernetesServiceOp) AddRegistry(ctx context.Context, req *KubernetesClusterRegistryRequest) (*Response, error) {
	path := fmt.Sprintf("%s/registry", kubernetesBasePath)
	request, err := svc.client.NewRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, err
	}
	resp, err := svc.client.Do(ctx, request, nil)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// RemoveRegistry removes docr registry support for all the specified clusters
func (svc *KubernetesServiceOp) RemoveRegistry(ctx context.Context, req *KubernetesClusterRegistryRequest) (*Response, error) {
	path := fmt.Sprintf("%s/registry", kubernetesBasePath)
	request, err := svc.client.NewRequest(ctx, http.MethodDelete, path, req)
	if err != nil {
		return nil, err
	}
	resp, err := svc.client.Do(ctx, request, nil)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

type runClusterlintRoot struct {
	RunID string `json:"run_id"`
}

// RunClusterlint schedules a clusterlint run for the specified cluster
func (svc *KubernetesServiceOp) RunClusterlint(ctx context.Context, clusterID string, req *KubernetesRunClusterlintRequest) (string, *Response, error) {
	path := fmt.Sprintf("%s/%s/clusterlint", kubernetesClustersPath, clusterID)
	request, err := svc.client.NewRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return "", nil, err
	}
	root := new(runClusterlintRoot)
	resp, err := svc.client.Do(ctx, request, root)
	if err != nil {
		return "", resp, err
	}
	return root.RunID, resp, nil
}

type clusterlintDiagnosticsRoot struct {
	Diagnostics []*ClusterlintDiagnostic
}

// GetClusterlintResults fetches the diagnostics after clusterlint run completes
func (svc *KubernetesServiceOp) GetClusterlintResults(ctx context.Context, clusterID string, req *KubernetesGetClusterlintRequest) ([]*ClusterlintDiagnostic, *Response, error) {
	path := fmt.Sprintf("%s/%s/clusterlint", kubernetesClustersPath, clusterID)
	if req != nil {
		v := make(url.Values)
		if req.RunId != "" {
			v.Set("run_id", req.RunId)
		}
		if query := v.Encode(); query != "" {
			path = path + "?" + query
		}
	}

	request, err := svc.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(clusterlintDiagnosticsRoot)
	resp, err := svc.client.Do(ctx, request, root)
	if err != nil {
		return nil, resp, err
	}
	return root.Diagnostics, resp, nil
}

func (svc *KubernetesServiceOp) GetClusterStatusMessages(ctx context.Context, clusterID string, req *KubernetesGetClusterStatusMessagesRequest) ([]*KubernetesClusterStatusMessage, *Response, error) {
	path := fmt.Sprintf("%s/%s/status_messages", kubernetesClustersPath, clusterID)

	if req != nil {
		v := make(url.Values)
		if req.Since != nil {
			v.Set("since", req.Since.Format(time.RFC3339))
		}
		if query := v.Encode(); query != "" {
			path = path + "?" + query
		}
	}

	request, err := svc.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(clusterStatusMessagesRoot)
	resp, err := svc.client.Do(ctx, request, root)
	if err != nil {
		return nil, resp, err
	}
	return root.Messages, resp, nil
}
