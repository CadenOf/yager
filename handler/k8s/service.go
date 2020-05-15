package k8s

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"

	. "voyager/handler"
	"voyager/model"
	SVC_CONST "voyager/pkg/constvar"
	"voyager/pkg/errno"
	"voyager/pkg/logger"
	"voyager/util"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//GetService get Service instances from k8s
func GetService(ctx *gin.Context) {
	log := logger.RuntimeLog
	zoneName := ctx.Param("zone")
	namespace := ctx.Param("ns")
	name := ctx.Param("name")

	// fetch k8s-client handler by zoneName
	kclient, err := GetClientByAzCode(zoneName)
	if err != nil {
		log.WithError(err)
		SendResponse(ctx, errno.ErrTokenInvalid, nil)
		return
	}

	startAt := time.Now()

	svc, err := kclient.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		SendResponse(ctx, err, "failed to get Service info.")
		return
	}
	logger.MetricsEmit(
		SVC_CONST.K8S_LOG_Method_GetService,
		util.GetReqID(ctx),
		float32(time.Since(startAt)/time.Millisecond),
		err == err,
	)
	SendResponse(ctx, errno.OK, svc)
}

//ListService List all Services from specify namespace
func ListService(ctx *gin.Context) {
	log := logger.RuntimeLog
	zoneName := ctx.Param("zone")
	namespace := ctx.Param("ns")

	// fetch k8s-client handler by zoneName
	kclient, err := GetClientByAzCode(zoneName)
	if err != nil {
		log.WithError(err)
		SendResponse(ctx, errno.ErrTokenInvalid, nil)
		return
	}

	startAt := time.Now()
	svcs, err := kclient.CoreV1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		SendResponse(ctx, err, "failed to get Service info.")
		return
	}
	logger.MetricsEmit(
		SVC_CONST.K8S_LOG_Method_ListService,
		util.GetReqID(ctx),
		float32(time.Since(startAt)/time.Millisecond),
		err == err,
	)

	SendResponse(ctx, errno.OK, svcs.Items)
}

// CreateService create Service instance
func CreateService(ctx *gin.Context) {
	log := logger.RuntimeLog
	var svcModel *model.Service
	if err := ctx.BindJSON(&svcModel); err != nil {
		SendResponse(ctx, err, "Request Body Invalid")
	}

	svcNamespace := strings.ToLower(svcModel.SvcMeta.Namespace)
	svcZone := svcModel.SvcMeta.AppMeta.ZoneName
	svcName := svcModel.SvcMeta.Name

	// fetch k8s-client hander by zoneName
	kclient, err := GetClientByAzCode(svcZone)
	if err != nil {
		log.WithError(err)
		SendResponse(ctx, errno.ErrTokenInvalid, nil)
		return
	}

	startAt := time.Now() // used to record operation time cost
	_, err = kclient.CoreV1().Services(svcNamespace).Create(makeupServiceData(ctx, svcModel))
	if err != nil {
		SendResponse(ctx, err, "create Service fail.")
		return
	}
	logger.MetricsEmit(
		SVC_CONST.K8S_LOG_Method_CreateService,
		util.GetReqID(ctx),
		float32(time.Since(startAt)/time.Millisecond),
		err == err,
	)
	SendResponse(ctx, errno.OK, fmt.Sprintf("Create Service %s success.", svcName))
}

// DeleteService delete Service instance
func DeleteService(ctx *gin.Context) {
	log := logger.RuntimeLog
	depZone := ctx.Param("zone")
	depNamespace := ctx.Param("ns")
	depName := ctx.Param("name")

	// fetch k8s-client handler by zoneName
	kclient, err := GetClientByAzCode(depZone)
	if err != nil {
		SendResponse(ctx, errno.ErrBind, nil)
		return
	}

	log.Info("Deleting Service...")
	deletePolicy := metav1.DeletePropagationForeground

	startAt := time.Now()
	err = kclient.CoreV1().Services(depNamespace).Delete(depName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	logger.MetricsEmit(
		SVC_CONST.K8S_LOG_Method_DeleteService,
		util.GetReqID(ctx),
		float32(time.Since(startAt)/time.Millisecond),
		err == nil || errors.IsNotFound(err),
	)
	if errors.IsNotFound(err) {
		log.Infof("Service %s not found in k8s", depName)
		SendResponse(ctx, err, nil)
		return
	}
	if err != nil {
		SendResponse(ctx, err, nil)
		return
	}

	delDepResData := fmt.Sprintf("Service %s success deleted.", depName)
	SendResponse(ctx, errno.OK, delDepResData)
}

// Makeup Service TemplateData
func makeupServiceData(ctx *gin.Context, svcModel *model.Service) *apiv1.Service {

	svcMeta := svcModel.SvcMeta
	svcSpec := svcModel.SvcSpec

	// init annotations
	annotations := map[string]string{
		SVC_CONST.K8S_RESOURCE_ANNOTATION_zone:  svcMeta.AppMeta.ZoneName,
		SVC_CONST.K8S_RESOURCE_ANNOTATION_orgid: svcMeta.AppMeta.OrgID,
		SVC_CONST.K8S_RESOURCE_ANNOTATION_appid: svcMeta.AppMeta.AppID,
		SVC_CONST.K8S_RESOURCE_ANNOTATION_env:   svcMeta.AppMeta.Env,
	}

	// init service ports
	var portTerms []apiv1.ServicePort
	for _, svcPorts := range svcSpec.Ports {
		portTerms = append(portTerms, apiv1.ServicePort{
			Protocol: apiv1.ProtocolTCP,
			Port:     svcPorts.Port,
			TargetPort: intstr.IntOrString{
				IntVal: svcPorts.TargetPort,
			},
			NodePort: svcPorts.NodePort,
		})

	}

	// init overall subset
	svcSet := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        svcMeta.Name,
			Annotations: annotations,
			Labels: map[string]string{
				SVC_CONST.K8S_RESOURCE_ANNOTATION_appid: svcMeta.AppMeta.AppID,
			},
		},
		Spec: apiv1.ServiceSpec{
			Selector: svcSpec.Selector,
			Ports:    portTerms,
			Type:     apiv1.ServiceType(svcSpec.Type),
		},
	}

	return svcSet
}
