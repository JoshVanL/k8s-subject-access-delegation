package subject_access_delegation

import (
	"fmt"
	"time"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger"
)

var (
	triggerUID = 1
)

func (s *SubjectAccessDelegation) BuildDeletionTriggers() error {
	for i := range s.sad.Spec.DeletionTriggers {
		s.sad.Spec.DeletionTriggers[i].UID = triggerUID
		triggerUID++
	}

	if err := s.updateRemoteSAD(); err != nil {
		return fmt.Errorf("failed to update trigger status against API server: %v", err)
	}

	triggers, err := trigger.New(s, s.sad.Spec.DeletionTriggers)
	if err != nil {
		return fmt.Errorf("failed to build deletion triggers: %v", err)
	}

	s.deletionTriggers = triggers
	return nil
}

func (s *SubjectAccessDelegation) BuildTriggers() error {
	for i := range s.sad.Spec.EventTriggers {
		s.sad.Spec.EventTriggers[i].UID = triggerUID
		triggerUID++
	}

	if err := s.updateRemoteSAD(); err != nil {
		return fmt.Errorf("failed to update trigger status against API server: %v", err)
	}

	triggers, err := trigger.New(s, s.sad.Spec.EventTriggers)
	if err != nil {
		return fmt.Errorf("failed to build triggers: %v", err)
	}

	s.triggers = triggers
	return nil
}

func (s *SubjectAccessDelegation) ActivateDeletionTriggers() (bool, error) {
	s.log.Debug("Activating Deletion Triggers")

	if err := s.updateLocalSAD(); err != nil {
		return false, err
	}

	allFired := true
	for _, trigger := range s.deletionTriggers {
		if !trigger.Completed() {
			trigger.Activate()
			allFired = false
		}
	}

	if allFired {
		s.log.Infof("All deletion triggers already triggered.")
		if err := s.cleanUpBindings(); err != nil {
			s.log.Errorf("Failed to clean up any remaining bingings: %v", err)
		}

		if err := s.updateTimeFired(0); err != nil {
			s.log.Errorf("Failed to update API server with 0 Activated Time: %v", err)
		}

		return false, nil
	}

	s.log.Info("Deletion Triggers Activated")

	if err := s.updateTimeFired(time.Now().Unix()); err != nil {
		s.log.Errorf("Failed to update API server with non-zero Activated Time: %v", err)
	}

	ready := false

	for !ready {
		if s.waitOnDeletionTriggers() {
			return true, nil
		}

		s.log.Info("All deletion triggers have been satisfied, checking still true")

		ready = s.checkDeletionTriggers()
		if !ready {
			s.log.Info("Not all deletion triggers ready at the same time, re-waiting.")
		}
	}

	if err := s.updateTimeFired(0); err != nil {
		s.log.Errorf("Failed to update API server with 0 Activated Time: %v", err)
	}

	s.log.Infof("All deletion triggers fired!")

	return false, nil
}

func (s *SubjectAccessDelegation) ActivateTriggers() (closed bool, err error) {
	s.log.Debug("Activating Triggers")

	if err := s.updateLocalSAD(); err != nil {
		return false, err
	}

	allFired := true
	for _, trigger := range s.triggers {
		if !trigger.Completed() {
			trigger.Activate()
			allFired = false
		}
	}

	if allFired {
		s.triggered = true

		s.log.Infof("All triggers already triggered.")
		if err := s.cleanUpBindings(); err != nil {
			s.log.Errorf("Failed to clean up any remaining bingings: %v", err)
		}

		if err := s.updateTimeActivated(0); err != nil {
			s.log.Errorf("Failed to update API server with 0 Activated Time: %v", err)
		}

		return false, nil
	}

	s.log.Info("Triggers Activated")

	if err := s.updateTimeActivated(time.Now().Unix()); err != nil {
		s.log.Errorf("Failed to update API server with non-zero Activated Time: %v", err)
	}

	ready := false

	for !ready {
		if s.waitOnTriggers() {
			return true, nil
		}

		s.log.Info("All triggers have been satisfied, checking still true")

		ready = s.checkTriggers()
		if !ready {
			s.log.Info("Not all triggers ready at the same time, re-waiting.")
		}
	}

	if err := s.updateTimeActivated(0); err != nil {
		s.log.Errorf("Failed to update API server with 0 Activated Time: %v", err)
	}

	s.log.Infof("All triggers fired!")
	s.triggered = true

	return false, nil
}

func (s *SubjectAccessDelegation) waitOnTriggers() (closed bool) {
	for _, trigger := range s.triggers {
		if trigger.WaitOn() {
			return true
		}
	}

	return false
}

func (s *SubjectAccessDelegation) waitOnDeletionTriggers() (closed bool) {
	for _, trigger := range s.deletionTriggers {
		if trigger.WaitOn() {
			return true
		}
	}

	return false
}

func (s *SubjectAccessDelegation) checkTriggers() (ready bool) {
	for _, trigger := range s.triggers {
		ready := trigger.Completed()
		if !ready {
			return false
		}
	}

	return true
}

func (s *SubjectAccessDelegation) checkDeletionTriggers() (ready bool) {
	for _, trigger := range s.deletionTriggers {
		ready := trigger.Completed()
		if !ready {
			return false
		}
	}

	return true
}

func (s *SubjectAccessDelegation) UpdateTriggerFired(uid int, fired bool) error {
	if err := s.updateLocalSAD(); err != nil {
		return err
	}

	found := false

	for i, trigger := range s.sad.Spec.EventTriggers {
		if trigger.UID == uid {
			s.sad.Spec.EventTriggers[i].Triggered = fired

			found = true
			break
		}
	}

	if !found {
		for i, trigger := range s.sad.Spec.DeletionTriggers {
			if trigger.UID == uid {
				s.sad.Spec.DeletionTriggers[i].Triggered = fired

				found = true
				break
			}
		}
	}

	if !found {
		return fmt.Errorf("failed to find trigger with SAD UID: %d", uid)
	}

	if err := s.updateRemoteSAD(); err != nil {
		return fmt.Errorf("failed to update trigger status against API server: %v", err)
	}

	return nil
}

func (s *SubjectAccessDelegation) updateTimeActivated(unixTime int64) error {
	if err := s.updateLocalSAD(); err != nil {
		return err
	}

	s.sad.Status.TimeActivated = unixTime

	if err := s.updateRemoteSAD(); err != nil {
		return fmt.Errorf("failed to update trigger activated time against API server: %v", err)
	}

	return nil
}

func (s *SubjectAccessDelegation) updateTimeFired(unixTime int64) error {
	if err := s.updateLocalSAD(); err != nil {
		return err
	}

	s.sad.Status.TimeFired = unixTime

	if err := s.updateRemoteSAD(); err != nil {
		return fmt.Errorf("failed to update trigger activated time against API server: %v", err)
	}

	return nil
}
