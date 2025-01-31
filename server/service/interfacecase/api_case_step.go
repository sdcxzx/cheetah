package interfacecase

import (
	"github.com/test-instructor/cheetah/server/global"
	"github.com/test-instructor/cheetah/server/model/common/request"
	"github.com/test-instructor/cheetah/server/model/interfacecase"
	interfacecaseReq "github.com/test-instructor/cheetah/server/model/interfacecase/request"
	"gorm.io/gorm"
)

type TestCaseService struct {
}

// CreateTestCase 创建TestCase记录

func (apicaseService *TestCaseService) CreateTestCaseStep(apicase interfacecase.ApiCaseStep) (err error) {
	err = global.GVA_DB.Create(&apicase).Error
	return err
}

// DeleteTestCase 删除TestCase记录

func (apicaseService *TestCaseService) DeleteTestCaseStep(apicase interfacecase.ApiCaseStep) (err error) {
	err = global.GVA_DB.Delete(&apicase).Error
	return err
}

// DeleteTestCaseByIds 批量删除TestCase记录

func (apicaseService *TestCaseService) DeleteTestCaseStepByIds(ids request.IdsReq) (err error) {
	err = global.GVA_DB.Delete(&[]interfacecase.ApiCaseStep{}, "id in ?", ids.Ids).Error
	return err
}

// UpdateTestCase 更新TestCase记录

func (apicaseService *TestCaseService) UpdateTestCaseStep(apicase interfacecase.ApiCaseStep) (err error) {
	var oId interfacecase.Operator
	global.GVA_DB.Model(interfacecase.ApiCaseStep{}).Where("id = ?", apicase.ID).First(&oId)
	apicase.CreatedByID = oId.CreatedByID
	apicase.TStep = []interfacecase.ApiStep{}
	err = global.GVA_DB.Save(&apicase).Error
	return err
}

// UpdateTestCase TestCase排序

func (apicaseService *TestCaseService) SortTestCaseStep(apicase interfacecase.ApiCaseStep) (err error) {

	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		for _, v := range apicase.TStep {
			err := tx.Find(&interfacecase.ApiStep{
				GVA_MODEL: global.GVA_MODEL{ID: v.ID},
			}).Update("Sort", v.Sort).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// AddTestCase TestCase排序

func (apicaseService *TestCaseService) AddTestCaseStep(apiCaseID request.ApiCaseIdReq) (caseApiDetail interfacecase.ApiStep, err error) {
	caseApiDetail = interfacecase.ApiStep{GVA_MODEL: global.GVA_MODEL{ID: apiCaseID.ApiID}}
	err = global.GVA_DB.Preload("Request").First(&caseApiDetail).Error
	if err != nil {
		return interfacecase.ApiStep{}, err
	}
	caseApiDetail.Parent = caseApiDetail.ID
	caseApiDetail.ID = 0
	caseApiDetail.Request.ID = 0
	caseApiDetail.ApiType = 2
	caseDetail := interfacecase.ApiCaseStep{GVA_MODEL: global.GVA_MODEL{ID: apiCaseID.CaseID}}
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var err error
		err = tx.Create(&caseApiDetail).Error
		if err != nil {
			return err
		}
		err = tx.Model(&caseDetail).Association("TStep").Append(&caseApiDetail)
		if err != nil {
			return err
		}
		return err
	})
	if err != nil {
		return interfacecase.ApiStep{}, err
	}
	return
}

// DelTestCase

func (apicaseService *TestCaseService) DelTestCaseStep(apiCaseID request.ApiCaseIdReq) (err error) {
	caseApiDetail := interfacecase.ApiStep{GVA_MODEL: global.GVA_MODEL{ID: apiCaseID.ApiID}}
	//err = global.GVA_DB.First(&caseApiDetail).Error
	//if err != nil {
	//	return err
	//}
	caseDetail := interfacecase.ApiCaseStep{GVA_MODEL: global.GVA_MODEL{ID: apiCaseID.CaseID}}
	err = global.GVA_DB.Model(&caseDetail).Association("TStep").Delete(&caseApiDetail)
	return
}

type ToTestCase struct {
	Config    interfacecase.ApiConfig
	TestSteps []interfacecase.ApiStep
}

// GetTestCase 根据id获取TestCase记录

func (apicaseService *TestCaseService) FindTestCaseStep(id uint) (err error, apicase interfacecase.ApiCaseStep) {
	err = global.GVA_DB.
		Preload("Project").
		Preload("TStep", func(db2 *gorm.DB) *gorm.DB {
			return db2.Order("Sort")
		}).
		Preload("TStep.Request").
		Where("id = ?", id).First(&apicase).Error
	return
}

// GetTestCaseInfoList 分页获取TestCase记录

func (apicaseService *TestCaseService) GetTestCaseStepInfoList(info interfacecaseReq.TestCaseSearch) (err error, list interface{}, total int64) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	// 创建db
	db := global.GVA_DB.Model(&interfacecase.ApiCaseStep{}).
		Preload("Project").Joins("Project").Where("Project.ID = ?", info.ProjectID)
	if info.ApiMenuID > 0 {
		db.Preload("ApiMenu").Joins("ApiMenu").Where("ApiMenu.ID = ?", info.ApiMenuID)
	}
	db.Where("type = ?", info.ApiType)
	var apicases []interfacecase.ApiCaseStep
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.Name != "" {
		db = db.Where("name LIKE ?", "%"+info.Name+"%")
	}
	if info.FrontCase {
		db.Where("front_case = ?", 1)
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}
	err = db.Preload("Project").Limit(limit).Offset(offset).Find(&apicases).Error
	return err, apicases, total
}
