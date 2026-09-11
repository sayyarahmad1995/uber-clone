package httpapi

import (
    "encoding/json"
    "errors"
    "net/http"
    "github.com/google/uuid"
    "github.com/sayyarahmad1995/uber-clone/backend/internal/driver"
)

func (api *API) getDriverSelection(w http.ResponseWriter,r *http.Request) {
    u,ok:=api.requireDriverCapability(w,r)
    if !ok { return }
    state,err:=api.drivers.OperatingState(r.Context(),u.ID)
    writeOperatingState(w,state,err)
}
func (api *API) setDriverSelection(w http.ResponseWriter,r *http.Request) {
    u,ok:=api.requireDriverCapability(w,r)
    if !ok { return }
    var body struct { VehicleID uuid.UUID `json:"vehicle_id"`; ServiceCode string `json:"service_code"` }
    if err:=json.NewDecoder(r.Body).Decode(&body); err!=nil {
        writeJSON(w,http.StatusBadRequest,map[string]string{"error":"invalid operating selection"})
        return
    }
    state,err:=api.drivers.SelectOperation(r.Context(),u.ID,body.VehicleID,body.ServiceCode)
    writeOperatingState(w,state,err)
}
func writeOperatingState(w http.ResponseWriter,state driver.OperatingState,err error) {
    switch {
    case errors.Is(err,driver.ErrNotFound):
        writeJSON(w,http.StatusNotFound,map[string]string{"error":"driver profile not found"})
    case errors.Is(err,driver.ErrSelectionInvalid),errors.Is(err,driver.ErrSelectionLocked):
        writeJSON(w,http.StatusConflict,map[string]string{"error":err.Error()})
    case err!=nil:
        writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"unable to load or save operating selection"})
    default:
        writeJSON(w,http.StatusOK,state)
    }
}
