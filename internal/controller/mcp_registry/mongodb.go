package mcp_registry

import (
	"context"
	"fmt"
	"log"
	"os"

	"errors"

	mcpv1alpha1 "github.com/RHEcosystemAppEng/mcp-registry-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/mcp-registry-operator/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	ErrNotFound       = errors.New("record not found")
	ErrAlreadyExists  = errors.New("record already exists")
	ErrInvalidInput   = errors.New("invalid input")
	ErrDatabase       = errors.New("database error")
	ErrInvalidVersion = errors.New("invalid version: cannot publish older version after newer version")
)

func CreateMongoDBResources(ctx context.Context, c client.Client, scheme *runtime.Scheme, mcpRegistry *mcpv1alpha1.McpRegistry) error {
	name := mcpRegistry.Name
	namespace := mcpRegistry.Namespace

	mongoLabels := map[string]string{"app": name, "component": "mongodb"}
	mongoPVC := &corev1.PersistentVolumeClaim{}
	mongoPVCName := name + "-mongodb-pvc"
	if err := c.Get(ctx, k8stypes.NamespacedName{Name: mongoPVCName, Namespace: namespace}, mongoPVC); apierrors.IsNotFound(err) {
		mongoPVC = &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      mongoPVCName,
				Namespace: namespace,
				Labels:    mongoLabels,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("500Mi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("500Mi"),
					},
				},
			},
		}
		if err := controllerutil.SetControllerReference(mcpRegistry, mongoPVC, scheme); err != nil {
			return fmt.Errorf("failed to set owner reference on mongodb pvc: %w", err)
		}
		if err := c.Create(ctx, mongoPVC); err != nil {
			return fmt.Errorf("failed to create mongodb pvc: %w", err)
		}
	} else if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	mongoDep := &appsv1.Deployment{}
	mongoDepName := name + "-mongodb"
	if err := c.Get(ctx, k8stypes.NamespacedName{Name: mongoDepName, Namespace: namespace}, mongoDep); apierrors.IsNotFound(err) {
		mongoDep = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      mongoDepName,
				Namespace: namespace,
				Labels:    mongoLabels,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(1),
				Selector: &metav1.LabelSelector{MatchLabels: mongoLabels},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: mongoLabels},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:  "mongodb",
							Image: "mongo:5.0",
							Ports: []corev1.ContainerPort{{ContainerPort: 27017}},
							VolumeMounts: []corev1.VolumeMount{{
								Name:      "mongodb-data",
								MountPath: "/data/db",
							}},
						}},
						Volumes: []corev1.Volume{{
							Name: "mongodb-data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: mongoPVCName,
								},
							},
						}},
					},
				},
			},
		}
		if err := controllerutil.SetControllerReference(mcpRegistry, mongoDep, scheme); err != nil {
			return fmt.Errorf("failed to set owner reference on mongodb deployment: %w", err)
		}
		if err := c.Create(ctx, mongoDep); err != nil {
			return fmt.Errorf("failed to create mongodb deployment: %w", err)
		}
	} else if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	mongoSvc := &corev1.Service{}
	mongoSvcName := GetMongoServiceName(name)
	if err := c.Get(ctx, k8stypes.NamespacedName{Name: mongoSvcName, Namespace: namespace}, mongoSvc); apierrors.IsNotFound(err) {
		mongoSvc = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      mongoSvcName,
				Namespace: namespace,
				Labels:    mongoLabels,
			},
			Spec: corev1.ServiceSpec{
				Selector: mongoLabels,
				Ports: []corev1.ServicePort{{
					Name:       "mongodb",
					Port:       27017,
					TargetPort: intstr.FromInt(27017),
				}},
				Type: corev1.ServiceTypeClusterIP,
			},
		}
		if err := controllerutil.SetControllerReference(mcpRegistry, mongoSvc, scheme); err != nil {
			return fmt.Errorf("failed to set owner reference on mongodb service: %w", err)
		}
		if err := c.Create(ctx, mongoSvc); err != nil {
			return fmt.Errorf("failed to create mongodb service: %w", err)
		}
	} else if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

// GetMongoServiceName returns the MongoDB service name for a given registry name
func GetMongoServiceName(registryName string) string {
	return registryName + "-mongodb"
}

type MongoDB struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

// String implements the fmt.Stringer interface to print MongoDB instance details.
func (m *MongoDB) String() string {
	if m == nil {
		return "<nil MongoDB instance>"
	}
	dbName := "<nil>"
	collName := "<nil>"
	if m.database != nil {
		dbName = m.database.Name()
	}
	if m.collection != nil {
		collName = m.collection.Name()
	}
	return fmt.Sprintf("MongoDB{Database: %s, Collection: %s, Client: %p}", dbName, collName, m.client)
}

func NewMongoDB(ctx context.Context, mcpServerRun *mcpv1alpha1.McpServerRun) (*MongoDB, error) {
	mcpRegistryName := mcpServerRun.Spec.RegistryRef.Name
	mcpRegistryNamespace := mcpServerRun.Spec.RegistryRef.Namespace
	if mcpRegistryNamespace == nil || *mcpRegistryNamespace == "" {
		mcpRegistryNamespace = &mcpServerRun.Namespace
	}

	var connectionURI string
	if os.Getenv("DEBUG_CONTROLLER") == "" {
		connectionURI = fmt.Sprintf("mongodb://%s.%s.svc.cluster.local:%s", GetMongoServiceName(mcpRegistryName), *mcpRegistryNamespace, types.MongoDBPort)
	} else {
		connectionURI = fmt.Sprintf("mongodb://localhost:%s", types.MongoDBPort)
	}
	fmt.Printf("connectionURI: %s\n", connectionURI)
	databaseName := types.MongoDBDatabaseName
	collectionName := types.MongoDBCollectionName

	// Set client options and connect to MongoDB
	clientOptions := options.Client().ApplyURI(connectionURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		fmt.Printf("Connection failed: %s\n", err)
		return nil, err
	}

	// Ping the MongoDB server to verify the connection
	if err = client.Ping(ctx, nil); err != nil {
		fmt.Printf("Ping failed: %s\n", err)
		return nil, err
	}

	// Get database and collection
	database := client.Database(databaseName)
	fmt.Printf("Got db %s", databaseName)
	collection := database.Collection(collectionName)
	fmt.Printf("Got collection %s", collectionName)

	// Create indexes for better query performance
	models := []mongo.IndexModel{
		{
			Keys: bson.D{bson.E{Key: "name", Value: 1}},
		},
		{
			Keys:    bson.D{bson.E{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		// add an index for the combination of name and version
		{
			Keys:    bson.D{bson.E{Key: "name", Value: 1}, bson.E{Key: "versiondetail.version", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err = collection.Indexes().CreateMany(ctx, models)
	if err != nil {
		fmt.Printf("Error creating indexes: %s\n", err)
		// Mongo will error if the index already exists, we can ignore this and continue.
		var commandError mongo.CommandError
		if errors.As(err, &commandError) && commandError.Code != 86 {
			return nil, err
		}
		log.Printf("Indexes already exists, skipping.")
	}

	return &MongoDB{
		client:     client,
		database:   database,
		collection: collection,
	}, nil
}

func InsertServerDetail(ctx context.Context, db *MongoDB, serverDetail *mcpv1alpha1.ServerDetail) error {
	// Find all elements in the collection and print them
	filter := bson.M{
		"name": serverDetail.Name,
	}

	var existingEntry mcpv1alpha1.ServerDetail
	err := db.collection.FindOne(ctx, filter).Decode(&existingEntry)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return fmt.Errorf("error checking existing entry: %w", err)
	}
	fmt.Printf("existingEntry: %+v\n", existingEntry)

	if existingEntry.Server.ID != "" {
		fmt.Printf("updating existing entry %s\n", existingEntry.ID)
		// check that the current version is greater than the existing one
		// if serverDetail.VersionDetail.Version <= existingEntry.VersionDetail.Version {
		// 	return fmt.Errorf("version must be greater than existing version")
		// }
		_, err = db.collection.UpdateOne(
			ctx,
			bson.M{"id": existingEntry.ID},
			bson.M{"$set": bson.M{"versiondetail.islatest": false}})
		if err != nil {
			return fmt.Errorf("error updating existing entry: %w", err)
		}
	} else {
		// Insert the entry into the database
		fmt.Printf("inserting new entry %s: %s\n", serverDetail.ID, serverDetail.Name)
		_, err = db.collection.InsertOne(ctx, serverDetail)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				// return ErrAlreadyExists
				fmt.Printf("entry already exists, skipping\n")
				return nil
			}
			return fmt.Errorf("error inserting entry: %w", err)
		}
	}

	return nil
}
