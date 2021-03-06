package config

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/deploy/spindeploy/changedetector"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const SpinnakerConfigHashKey = "config"
const KustomizeHashKey = "kustomize"

type changeDetector struct {
	log         logr.Logger
	evtRecorder record.EventRecorder
}

type ChangeDetectorGenerator struct{}

func (g *ChangeDetectorGenerator) NewChangeDetector(client client.Client, log logr.Logger, evtRecorder record.EventRecorder, scheme *runtime.Scheme) (changedetector.ChangeDetector, error) {
	return &changeDetector{log: log, evtRecorder: evtRecorder}, nil
}

// IsSpinnakerUpToDate returns true if the Config has changed compared to the last recorded status hash
func (ch *changeDetector) IsSpinnakerUpToDate(ctx context.Context, spinSvc interfaces.SpinnakerService) (bool, error) {
	upd, err := ch.isUpToDate(spinSvc.GetSpinnakerConfig(), SpinnakerConfigHashKey, spinSvc)
	if err != nil {
		return false, err
	}

	kUpd, err := ch.isUpToDate(spinSvc.GetKustomization(), KustomizeHashKey, spinSvc)
	return upd && kUpd, err
}

func (ch *changeDetector) isUpToDate(config interface{}, hashKey string, spinSvc interfaces.SpinnakerService) (bool, error) {
	h, err := ch.getHash(config)
	if err != nil {
		return false, err
	}

	st := spinSvc.GetStatus()
	prior := st.UpdateHashIfNotExist(hashKey, h, time.Now())
	return h == prior.Hash, nil
}

func (ch *changeDetector) getHash(config interface{}) (string, error) {
	data, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	m := md5.Sum(data)
	return hex.EncodeToString(m[:]), nil
}

func (ch *changeDetector) AlwaysRun() bool {
	return true
}
