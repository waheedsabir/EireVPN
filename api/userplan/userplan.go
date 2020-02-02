package userplan

import (
	"eirevpn/api/errors"
	"eirevpn/api/logger"
	"eirevpn/api/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// UserPlan fetches a plan by ID
func UserPlan(c *gin.Context) {
	userPlanID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var userplan models.UserPlan
	userplan.ID = uint(userPlanID)
	if err := userplan.Find(); err != nil {
		logger.Log(logger.Fields{
			Loc:   "/userplans/:id - UserPlan()",
			Code:  errors.UserPlanNotFound.Code,
			Extra: map[string]interface{}{"UserPlanID": c.Param("id")},
			Err:   err.Error(),
		})
		c.AbortWithStatusJSON(errors.UserPlanNotFound.Status, errors.UserPlanNotFound)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": 200,
		"errors": make([]string, 0),
		"data": gin.H{
			"userplan": userplan,
		},
	})

}

// CreateUserPlan creates a new plan
func CreateUserPlan(c *gin.Context) {
	var userplan models.UserPlan
	type UserPlanCreate struct {
		UserID     uint   `json:"user_id" binding:"required"`
		PlanID     uint   `json:"plan_id" binding:"required"`
		Active     string `json:"active" binding:"required"`
		StartDate  string `json:"start_date" binding:"required"`
		ExpiryDate string `json:"expiry_date" binding:"required"`
	}
	userPlanCreate := UserPlanCreate{}
	if err := c.BindJSON(&userPlanCreate); err != nil {
		logger.Log(logger.Fields{
			Loc:  "/userplans/create - CreateUserPlan()",
			Code: errors.InvalidForm.Code,
			Err:  err.Error(),
		})
		c.AbortWithStatusJSON(errors.InvalidForm.Status, errors.InvalidForm)
		return
	}

	userplan.UserID = userPlanCreate.UserID
	userplan.PlanID = userPlanCreate.PlanID
	userplan.Active = userPlanCreate.Active == "true"
	startdate, _ := time.Parse("2006-01-02 15:04", userPlanCreate.StartDate)
	userplan.StartDate = startdate
	expirydate, _ := time.Parse("2006-01-02 15:04", userPlanCreate.ExpiryDate)
	userplan.ExpiryDate = expirydate

	if err := userplan.Create(); err != nil {
		logger.Log(logger.Fields{
			Loc:  "/userplans/create - CreateUserPlan()",
			Code: errors.InternalServerError.Code,
			Extra: map[string]interface{}{
				"UserPlan": userplan,
			},
			Err: err.Error(),
		})
		c.AbortWithStatusJSON(errors.InternalServerError.Status, errors.InternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": 200,
		"data": gin.H{
			"userplan": userplan,
		},
	})
}

// DeleteUserPlan deletes a given users userplan. It will not delete a userplan fully however,
// it will just set a DeletedAt datetime on the record
func DeleteUserPlan(c *gin.Context) {
	userPlanID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var userplan models.UserPlan
	userplan.ID = uint(userPlanID)
	if err := userplan.Find(); err != nil {
		logger.Log(logger.Fields{
			Loc:   "/userplans/delete/:id - DeleteUserPlan()",
			Code:  errors.UserPlanNotFound.Code,
			Extra: map[string]interface{}{"UserPlanID": c.Param("id")},
			Err:   err.Error(),
		})
		c.AbortWithStatusJSON(errors.UserPlanNotFound.Status, errors.UserPlanNotFound)
		return
	}

	if err := userplan.Delete(); err != nil {
		logger.Log(logger.Fields{
			Loc:   "/userplans/delete/:id - DeleteUserPlan()",
			Code:  errors.PlanNotFound.Code,
			Extra: map[string]interface{}{"UserPlanID": c.Param("id")},
			Err:   err.Error(),
		})
		c.AbortWithStatusJSON(errors.InternalServerError.Status, errors.InternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": 200,
	})
}

// UpdateUserPlan updates an existing plan
func UpdateUserPlan(c *gin.Context) {
	userPlanID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var userplan models.UserPlan
	userplan.ID = uint(userPlanID)

	type UserPlanUpdates struct {
		Active     string `json:"active" binding:"required"`
		StartDate  string `json:"start_date" binding:"required"`
		ExpiryDate string `json:"expiry_date" binding:"required"`
	}
	userPlanUdates := UserPlanUpdates{}

	if err := userplan.Find(); err != nil {
		logger.Log(logger.Fields{
			Loc:   "/userplans/update/:id - UpdateUserPlan()",
			Code:  errors.UserPlanNotFound.Code,
			Extra: map[string]interface{}{"UserPlanID": c.Param("id")},
			Err:   err.Error(),
		})
		c.AbortWithStatusJSON(errors.UserPlanNotFound.Status, errors.UserPlanNotFound)
		return
	}

	if err := c.BindJSON(&userPlanUdates); err != nil {
		logger.Log(logger.Fields{
			Loc:   "/userplans/update/:id - UpdateUserPlan()",
			Code:  errors.InvalidForm.Code,
			Extra: map[string]interface{}{"UserPlanID": c.Param("id")},
			Err:   err.Error(),
		})
		c.AbortWithStatusJSON(errors.InvalidForm.Status, errors.InvalidForm)
		return
	}

	userplan.Active = userPlanUdates.Active == "true"
	startdate, _ := time.Parse("2006-01-02 15:04", userPlanUdates.StartDate)
	userplan.StartDate = startdate
	expirydate, _ := time.Parse("2006-01-02 15:04", userPlanUdates.ExpiryDate)
	userplan.ExpiryDate = expirydate
	if err := userplan.Save(); err != nil {
		logger.Log(logger.Fields{
			Loc:   "/userplans/update/:id - UpdateUserPlan()",
			Code:  errors.InternalServerError.Code,
			Extra: map[string]interface{}{"UserPlanID": c.Param("id")},
			Err:   err.Error(),
		})
		c.AbortWithStatusJSON(errors.InternalServerError.Status, errors.InternalServerError)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": 200,
	})
}

// AllUserPlans returns an array of all available plans
func AllUserPlans(c *gin.Context) {
	var plans models.AllUserPlans

	if err := plans.FindAll(); err != nil {
		logger.Log(logger.Fields{
			Loc:  "/userplans - AllUserPlans()",
			Code: errors.InternalServerError.Code,
			Err:  err.Error(),
		})
		c.AbortWithStatusJSON(errors.InternalServerError.Status, errors.InternalServerError)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": 200,
		"data": gin.H{
			"plans": plans,
		},
	})
}
