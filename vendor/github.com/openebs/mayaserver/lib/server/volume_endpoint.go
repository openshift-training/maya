package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/openebs/mayaserver/lib/api/v1"
	"github.com/openebs/mayaserver/lib/volumeprovisioner"
)

// VSMSpecificRequest is a http handler implementation. It deals with HTTP
// requests w.r.t a single VSM.
//
// TODO
//    Should it return specific types than interface{} ?
func (s *HTTPServer) VSMSpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {

	s.logger.Printf("[DEBUG] Processing %v request", req.Method)
	switch req.Method {
	case "PUT", "POST":
		return s.vsmAdd(resp, req)
	case "GET":
		return s.vsmSpecificGetRequest(resp, req)
	default:
		return nil, CodedError(405, ErrInvalidMethod)
	}
}

// vsmSpecificGetRequest deals with HTTP GET request w.r.t a single VSM
func (s *HTTPServer) vsmSpecificGetRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	// Extract info from path after trimming
	path := strings.TrimPrefix(req.URL.Path, "/latest/volumes")

	// Is req valid ?
	if path == req.URL.Path {
		return nil, CodedError(405, ErrInvalidMethod)
	}

	switch {

	case strings.Contains(path, "/info/"):
		vsmName := strings.TrimPrefix(path, "/info/")
		return s.vsmRead(resp, req, vsmName)
	case strings.Contains(path, "/delete/"):
		vsmName := strings.TrimPrefix(path, "/delete/")
		return s.vsmDelete(resp, req, vsmName)
	case path == "/":
		return s.vsmList(resp, req)
	default:
		return nil, CodedError(405, ErrInvalidMethod)
	}
}

// vsmList is the http handler that lists VSMs
func (s *HTTPServer) vsmList(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	s.logger.Printf("[DEBUG] Processing vsmList request")

	// Create a PVC
	pvc := &v1.PersistentVolumeClaim{}

	// Get the persistent volume provisioner instance
	pvp, err := volumeprovisioner.GetVolumeProvisioner(pvc.Labels)
	if err != nil {
		return nil, err
	}

	// Set the volume provisioner profile to provisioner
	_, err = pvp.Profile(pvc)
	if err != nil {
		return nil, err
	}

	lister, ok, err := pvp.Lister()
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("VSM list is not supported by '%s:%s'", pvp.Label(), pvp.Name())
	}

	l, err := lister.List()
	if err != nil {
		return nil, err
	}

	return l, nil
}

// vsmRead is the http handler that fetches the details of a VSM
func (s *HTTPServer) vsmRead(resp http.ResponseWriter, req *http.Request, vsmName string) (interface{}, error) {
	s.logger.Printf("[DEBUG] Processing vsmRead request")
	if vsmName == "" {
		return nil, fmt.Errorf("VSM name is missing")
	}

	// Create a PVC
	pvc := &v1.PersistentVolumeClaim{}
	pvc.Name = vsmName

	// Get persistent volume provisioner instance
	pvp, err := volumeprovisioner.GetVolumeProvisioner(pvc.Labels)
	if err != nil {
		return nil, err
	}

	// Set the volume provisioner profile to provisioner
	_, err = pvp.Profile(pvc)
	if err != nil {
		return nil, err
	}

	reader, ok := pvp.Reader()
	if !ok {
		return nil, fmt.Errorf("VSM read is not supported by '%s:%s'", pvp.Label(), pvp.Name())
	}

	// TODO
	// pvc should not be passed again !!
	details, err := reader.Read(pvc)
	if err != nil {
		return nil, err
	}

	return details, nil
}

// vsmDelete is the http handler that fetches the details of a VSM
func (s *HTTPServer) vsmDelete(resp http.ResponseWriter, req *http.Request, vsmName string) (interface{}, error) {
	if vsmName == "" {
		return nil, fmt.Errorf("VSM name is missing")
	}

	// Create a PVC
	pvc := &v1.PersistentVolumeClaim{}
	pvc.Name = vsmName

	// Get the persistent volume provisioner instance
	pvp, err := volumeprovisioner.GetVolumeProvisioner(pvc.Labels)
	if err != nil {
		return nil, err
	}

	// Set the volume provisioner profile to provisioner
	_, err = pvp.Profile(pvc)
	if err != nil {
		return nil, err
	}

	remover, ok, err := pvp.Remover()
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("VSM delete is not supported by '%s:%s'", pvp.Label(), pvp.Name())
	}

	err = remover.Remove()
	if err != nil {
		return nil, err
	}

	return fmt.Sprintf("VSM '%s' deleted successfully", vsmName), nil
}

// vsmAdd is the http handler that fetches the details of a VSM
func (s *HTTPServer) vsmAdd(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	s.logger.Printf("[DEBUG] Processing vsmAdd request")
	pvc := v1.PersistentVolumeClaim{}

	// The yaml/json spec is decoded to pvc struct
	if err := decodeBody(req, &pvc); err != nil {
		return nil, CodedError(400, err.Error())
	}

	// Name is expected to be available even in the minimalist specs
	if pvc.Name == "" {
		return nil, CodedError(400, fmt.Sprintf("VSM name missing in '%v'", pvc))
	}

	// Get persistent volume provisioner instance
	pvp, err := volumeprovisioner.GetVolumeProvisioner(pvc.Labels)
	if err != nil {
		return nil, err
	}

	// Set the volume provisioner profile to provisioner
	_, err = pvp.Profile(&pvc)
	if err != nil {
		return nil, err
	}

	adder, ok := pvp.Adder()
	if !ok {
		return nil, fmt.Errorf("VSM add is not supported by '%s:%s'", pvp.Label(), pvp.Name())
	}

	// TODO
	// pvc should not be passed again !!
	details, err := adder.Add(&pvc)
	if err != nil {
		return nil, err
	}

	return details, nil
}
