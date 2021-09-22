package controllers

import (
	"context"
	"time"

	samplecontrollerv1alpha1 "github.com/kou164nkn/foo-controller/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// You shuould run `KUBEBUILDER_ASSETS=$(pwd)/testbin/bin go test ./controller/...` at project root
var _ = Describe("when no existing resource exit", func() {
	ctx := context.TODO()

	var ns *corev1.Namespace

	BeforeEach(func() {
		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: "testns-" + randStringRunes(5)},
		}

		err := k8sClient.Create(ctx, ns)
		Expect(err).ToNot(HaveOccurred(), "failed to create test namespace")
	})

	AfterEach(func() {
		err := k8sClient.Delete(ctx, ns)
		Expect(err).ToNot(HaveOccurred(), "failed to delete test namespace")
	})

	It("should create a new Deployment resource with the specified name and replica if one replicas is provided", func() {
		By("Creating new Foo resource")
		foo := &samplecontrollerv1alpha1.Foo{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example-foo",
				Namespace: ns.Name,
			},
			Spec: samplecontrollerv1alpha1.FooSpec{
				DeploymentName: "example-foo",
				Replicas:       pointer.Int32Ptr(1),
			},
		}

		err := k8sClient.Create(ctx, foo)
		Expect(err).ToNot(HaveOccurred(), "failed to create test Foo resource")

		By("Cheking new Deployment resource created")
		deployment := &appsv1.Deployment{}
		Eventually(
			getResourceFunc(ctx, client.ObjectKey{Name: "example-foo", Namespace: foo.Namespace}, deployment),
			time.Second*5, time.Millisecond*500,
		).Should(BeNil())

		Expect(*deployment.Spec.Replicas).To(Equal(int32(1)))
	})

	It("Should create a new Deployment resource with the specified name and two replicas if two is specified", func() {
		By("Creating new Foo resource")
		foo := &samplecontrollerv1alpha1.Foo{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example-foo",
				Namespace: ns.Name,
			},
			Spec: samplecontrollerv1alpha1.FooSpec{
				DeploymentName: "example-foo",
				Replicas:       pointer.Int32Ptr(2),
			},
		}

		err := k8sClient.Create(ctx, foo)
		Expect(err).ToNot(HaveOccurred(), "failed to create test Foo resource")

		By("Cheking new Deployment resource created")
		deployment := &appsv1.Deployment{}
		Eventually(
			getResourceFunc(ctx, client.ObjectKey{Name: "example-foo", Namespace: foo.Namespace}, deployment),
			time.Second*5, time.Millisecond*500,
		).Should(BeNil())

		Expect(*deployment.Spec.Replicas).To(Equal(int32(2)))
	})

	It("should allow updating the replicas count after creating a Foo resource", func() {
		deploymentObjectKey := client.ObjectKey{
			Name:      "example-foo",
			Namespace: ns.Name,
		}
		fooObjectKey := client.ObjectKey{
			Name:      "example-foo",
			Namespace: ns.Name,
		}
		foo := &samplecontrollerv1alpha1.Foo{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fooObjectKey.Name,
				Namespace: fooObjectKey.Namespace,
			},
			Spec: samplecontrollerv1alpha1.FooSpec{
				DeploymentName: deploymentObjectKey.Name,
				Replicas:       pointer.Int32Ptr(1),
			},
		}

		By("Creating new Foo resource")
		err := k8sClient.Create(ctx, foo)
		Expect(err).ToNot(HaveOccurred(), "failed to create test Foo resource")

		deployment := &appsv1.Deployment{}
		Eventually(
			getResourceFunc(ctx, deploymentObjectKey, deployment),
			time.Second*5, time.Millisecond*500,
		).Should(BeNil(), "deployment resource should exist")

		Expect(*deployment.Spec.Replicas).To(Equal(int32(1)), "replicas count should be equal 1")

		By("Updating replicas count in existing Foo resource")
		foo.Spec.Replicas = pointer.Int32Ptr(2)
		err = k8sClient.Update(ctx, foo)
		Expect(err).ToNot(HaveOccurred(), "failed to Update Foo resource")

		By("Cheking Deployment resource scaled")
		Eventually(
			getDeploymentReplicasFunc(ctx, deploymentObjectKey),
		).Should(Equal(int32(2)), "expected Deployment resource to be scale to 2 replicas")
	})

	It("Should clean up an old Deployment resource if the deploymentName is changed", func() {
		deploymentObjectKey := client.ObjectKey{
			Name:      "example-foo",
			Namespace: ns.Name,
		}
		newDeploymentObjectKey := client.ObjectKey{
			Name:      "example-foo-new",
			Namespace: ns.Name,
		}
		fooObjectKey := client.ObjectKey{
			Name:      "example-foo",
			Namespace: ns.Name,
		}
		foo := &samplecontrollerv1alpha1.Foo{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fooObjectKey.Name,
				Namespace: fooObjectKey.Namespace,
			},
			Spec: samplecontrollerv1alpha1.FooSpec{
				DeploymentName: deploymentObjectKey.Name,
				Replicas:       pointer.Int32Ptr(1),
			},
		}

		By("Creating new Foo resource")
		err := k8sClient.Create(ctx, foo)
		Expect(err).ToNot(HaveOccurred(), "failed to create test Foo resource")

		deployment := &appsv1.Deployment{}
		Eventually(
			getResourceFunc(ctx, deploymentObjectKey, deployment),
			time.Second*5, time.Millisecond*500,
		).Should(BeNil(), "deployment resource should exist")

		err = k8sClient.Get(ctx, fooObjectKey, foo)
		Expect(err).ToNot(HaveOccurred(), "failed to retrieve Foo resource")

		By("Updating deploymentName in existing Foo resource")
		foo.Spec.DeploymentName = newDeploymentObjectKey.Name
		err = k8sClient.Update(ctx, foo)
		Expect(err).ToNot(HaveOccurred(), "failed to Update Foo resource")

		By("Checking old deployment deleted and new deployment created")
		Eventually(
			getResourceFunc(ctx, deploymentObjectKey, deployment),
			time.Second*5, time.Millisecond*500,
		).ShouldNot(BeNil(), "old deployment resource should be deleted")

		Eventually(
			getResourceFunc(ctx, newDeploymentObjectKey, deployment),
			time.Second*5, time.Millisecond*500,
		).Should(BeNil(), "new deployment resource should be created")
	})
})

func getResourceFunc(ctx context.Context, key client.ObjectKey, obj client.Object) func() error {
	return func() error {
		return k8sClient.Get(ctx, key, obj)
	}
}

func getDeploymentReplicasFunc(ctx context.Context, key client.ObjectKey) func() int32 {
	return func() int32 {
		depl := &appsv1.Deployment{}
		err := k8sClient.Get(ctx, key, depl)
		Expect(err).ToNot(HaveOccurred(), "failed to get Deployment resource")

		return *depl.Spec.Replicas
	}
}
