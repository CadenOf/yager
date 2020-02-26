package k8s

import (
	"fmt"
	"time"

	. "yager/handler"
	"yager/pkg/errno"
	"yager/pkg/logger"
	"yager/util"

	apiv1 "k8s.io/api/core/v1"
	//"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"time"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/api/errors"
)

//GetPod get pod instance from k8s
func GetPod(ctx *gin.Context) {

	zoneName := ctx.Param("zone")
	namespace := ctx.Param("ns")
	name := ctx.Param("name")

	cs, err := GetClientByAzCode(zoneName)
	if err != nil {
		SendResponse(ctx, errno.ErrTokenInvalid, nil)
	}

	//namespace = "kube-system"
	//name = "prometheus-fcbfd5fb4-p6rkz"
	fmt.Print(zoneName, namespace, name, "\n")

	begin := time.Now()

	pod, err := cs.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
	logger.MetricsEmit(
		"k8s.get_pod",
		util.GetReqID(ctx),
		float32(time.Since(begin)/time.Millisecond),
		err == nil,
	)

	fmt.Printf("logInfo: %s\n", pod)
	SendResponse(ctx, errno.OK, pod)
}

//ListPodByLabel get pod instance
func ListPodByLabel(ctx *gin.Context, zoneName, namespace, label string) (*apiv1.PodList, error) {
	cs, err := GetClientByAzCode(zoneName)
	if err != nil {
		SendResponse(ctx, errno.ErrTokenInvalid, nil)
	}

	begin := time.Now()

	podList, err := cs.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: label})

	logger.MetricsEmit(
		"k8s.list_pod_by_label",
		util.GetReqID(ctx),
		float32(time.Since(begin)/time.Millisecond),
		err == nil,
	)

	return podList, err
}

//GetPodHostIP get host ip
func GetPodHostIP(ctx *gin.Context, zoneName, ns, name string) (string, error) {
	cs, err := GetClientByAzCode(zoneName)
	if err != nil {
		return "", err

	}

	begin := time.Now()
	pod, err := cs.CoreV1().Pods(ns).Get(name, metav1.GetOptions{})
	logger.MetricsEmit(
		"k8s.get_pod",
		util.GetReqID(ctx),
		float32(time.Since(begin)/time.Millisecond),
		err == nil || errors.IsNotFound(err),
	)

	if err != nil {
		return "", err
	}

	return pod.Status.HostIP, nil
}

//func GetPodStatus(ctx *gin.Context, zoneName, namespace, name string, specId int64) (*dto.PodStatus, error) {
//	cs, err := GetClientByAzCode(zoneName)
//	if err != nil {
//		return nil, err
//
//	}
//
//	begin := time.Now()
//	pod, err := cs.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
//	logger.MetricsEmit(
//		"k8s.get_pod",
//		util.GetReqID(ctx),
//		float32(time.Since(begin)/time.Millisecond),
//		err == nil || errors.IsNotFound(err),
//	)
//
//	if err != nil {
//		return nil, err
//	}
//
//	return StatusOfPod(pod, specId)
//}
//
//func StatusOfPod(pod *apiv1.Pod, specId int64) (status *dto.PodStatus, err error) {
//	// Get pod conditions
//	conditions := make([]dto.PodCondition, 0, 3)
//	var healthy bool
//	for _, c := range pod.Status.Conditions {
//		conditions = append(conditions, dto.PodCondition{
//			Type:    string(c.Type),
//			Status:  string(c.Status),
//			Message: c.Message,
//			Reason:  c.Reason,
//		})
//
//		if c.Type == apiv1.PodReady && c.Status == apiv1.ConditionTrue {
//			healthy = true
//		}
//	}
//
//	// Get pod current version(specid)
//	var id int64
//	id, err = strconv.ParseInt(pod.Annotations["app.cdos.ctrip.com/specid"], 10, 64)
//	if err != nil {
//		return
//	}
//
//	if pod.Status.Phase != apiv1.PodRunning {
//		healthy = false
//	}
//
//	status = &dto.PodStatus{
//		Phase:            string(pod.Status.Phase),
//		IsCurrentVersion: specId == id,
//		Image:            pod.Spec.Containers[0].Image,
//		Conditions:       conditions,
//		Healthy:          healthy,
//	}
//	return
//}
//
//func DeletePod(ctx *gin.Context, zoneName, ns, name string) error {
//	cs, err := GetClientByAzCode(zoneName)
//	if err != nil {
//		return err
//	}
//
//	deletePolicy := metav1.DeletePropagationForeground
//
//	begin := time.Now()
//	err = cs.CoreV1().Pods(ns).Delete(name, &metav1.DeleteOptions{
//		PropagationPolicy: &deletePolicy,
//	})
//	logger.MetricsEmit(
//		"k8s.delete_pod",
//		util.GetReqID(ctx),
//		float32(time.Since(begin)/time.Millisecond),
//		err == nil,
//	)
//
//	return err
//}
