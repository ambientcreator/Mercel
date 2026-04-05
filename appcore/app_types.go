package appcore

import (
	"context"
	"database/sql"
	"sync"
)

const (
	CategoryPrimary   = "primary"
	CategorySecondary = "secondary"
	CategoryClosing   = "closing"
	AppStorageDirName = "MercelData"
	AppStorageDBName  = "mercel.sqlite"

	RoleAdmin                     = "admin"
	RoleGlobalDirector            = "global_director"
	RoleExecutiveDirector         = "executive_director"
	RoleTechnicalDirector         = "technical_director"
	DepartmentGlobal              = "global"
	DepartmentSupport             = "support"
	DepartmentTechnical           = "technical"
	DepartmentTelecom             = "telecom"
	DepartmentSKUD                = "skud"
	DepartmentApproval            = "approval"
	DepartmentMarketing           = "marketing"
	DepartmentCommercial          = "commercial"
	DepartmentFinance             = "finance"
	DepartmentLegal               = "legal"
	DepartmentDevelopment         = "development"
	RoleSupportHead               = "support_head"
	RoleSupportSenior             = "support_senior"
	RoleSupportEmployee           = "support_employee"
	RoleSupportSysadmin           = "support_sysadmin"
	RoleTechnicalHead             = "technical_head"
	RoleTechnicalSenior           = "technical_senior"
	RoleTechnicalEmployee         = "technical_employee"
	RoleTelecomDirector           = "telecom_construction_director"
	RoleTelecomHead               = "telecom_construction_head"
	RoleTelecomSeniorVOLS         = "telecom_senior_vols"
	RoleTelecomSeniorLVS          = "telecom_senior_lvs"
	RoleTelecomEmployeeVOLS       = "telecom_employee_vols"
	RoleTelecomEmployeeLVS        = "telecom_employee_lvs"
	RoleSKUDHead                  = "skud_head"
	RoleSKUDProjectManager        = "skud_project_manager"
	RoleSKUDSeniorService         = "skud_senior_service_engineer"
	RoleSKUDSeniorInstaller       = "skud_senior_installer"
	RoleSKUDServiceEngineer       = "skud_service_engineer"
	RoleSKUDInstaller             = "skud_installer"
	RoleApprovalHead              = "approval_head"
	RoleApprovalSenior            = "approval_senior"
	RoleApprovalEmployee          = "approval_employee"
	RoleMarketingHead             = "marketing_head"
	RoleMarketingCourier          = "marketing_courier"
	RoleCommercialDirector        = "commercial_director"
	RoleCommercialSubscriberHead  = "commercial_subscriber_head"
	RoleCommercialSeniorMRK       = "commercial_senior_mrk"
	RoleCommercialSeniorMRYU      = "commercial_senior_mryu"
	RoleCommercialActiveSalesHead = "commercial_active_sales_head"
	RoleCommercialEmployeeMRK     = "commercial_employee_mrk"
	RoleCommercialEmployeeMRYU    = "commercial_employee_mryu"
	RoleFinanceHead               = "finance_head"
	RoleFinanceEmployee           = "finance_employee"
	RoleLegalEmployee             = "legal_employee"
	RoleDevelopmentHead           = "development_head"
	RoleDevelopmentSenior         = "development_senior"
	RoleDevelopmentEmployee       = "development_employee"

	RoleSupportManager          = RoleSupportHead
	RoleSupportSeniorSpecialist = RoleSupportSenior
	RoleTechManager             = RoleTechnicalHead
	RoleSeniorTech              = RoleTechnicalSenior
	RoleTechnician              = RoleTechnicalEmployee
	RoleMRKManager              = RoleCommercialSubscriberHead
	RoleSeniorMRK               = RoleCommercialSeniorMRK
	RoleMRKEmployee             = RoleCommercialEmployeeMRK
	RoleManager                 = RoleSupportHead
	RoleSeniorSpecialist        = RoleSupportSenior
	RoleEmployee                = RoleSupportEmployee
)

type roleMeta struct {
	Department string
	Level      int
	Label      string
}

var roleCatalog = map[string]roleMeta{
	RoleAdmin:                     {Department: DepartmentGlobal, Level: 5, Label: "Р В РЎвЂ™Р В РўвЂР В РЎВР В РЎвЂР В Р вЂ¦Р В РЎвЂР РЋР С“Р РЋРІР‚С™Р РЋР вЂљР В Р’В°Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљ"},
	RoleGlobalDirector:            {Department: DepartmentGlobal, Level: 4, Label: "Р В РІР‚СљР В Р’ВµР В Р вЂ¦Р В Р’ВµР РЋР вЂљР В Р’В°Р В Р’В»Р РЋР Р‰Р В Р вЂ¦Р РЋРІР‚в„–Р В РІвЂћвЂ“ Р В РўвЂР В РЎвЂР РЋР вЂљР В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В РЎвЂўР РЋР вЂљ"},
	RoleExecutiveDirector:         {Department: DepartmentGlobal, Level: 4, Label: "Р В Р’ВР РЋР С“Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В Р вЂ¦Р В РЎвЂР РЋРІР‚С™Р В Р’ВµР В Р’В»Р РЋР Р‰Р В Р вЂ¦Р РЋРІР‚в„–Р В РІвЂћвЂ“ Р В РўвЂР В РЎвЂР РЋР вЂљР В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В РЎвЂўР РЋР вЂљ"},
	RoleTechnicalDirector:         {Department: DepartmentGlobal, Level: 4, Label: "Р В РЎС›Р В Р’ВµР РЋРІР‚В¦Р В Р вЂ¦Р В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂР В РІвЂћвЂ“ Р В РўвЂР В РЎвЂР РЋР вЂљР В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В РЎвЂўР РЋР вЂљ"},
	RoleSupportHead:               {Department: DepartmentSupport, Level: 3, Label: "Р В Р’В Р РЋРЎвЂњР В РЎвЂќР В РЎвЂўР В Р вЂ Р В РЎвЂўР В РўвЂР В РЎвЂР РЋРІР‚С™Р В Р’ВµР В Р’В»Р РЋР Р‰ Р В РЎС›Р В Р’ВµР РЋРІР‚В¦. Р В РЎСџР В РЎвЂўР В РўвЂР В РўвЂР В Р’ВµР РЋР вЂљР В Р’В¶Р В РЎвЂќР В РЎвЂ"},
	RoleSupportSysadmin:           {Department: DepartmentSupport, Level: 3, Label: "Р В Р Р‹Р В РЎвЂР РЋР С“Р РЋРІР‚С™Р В Р’ВµР В РЎВР В Р вЂ¦Р РЋРІР‚в„–Р В РІвЂћвЂ“ Р В Р’В°Р В РўвЂР В РЎВР В РЎвЂР В Р вЂ¦Р В РЎвЂР РЋР С“Р РЋРІР‚С™Р РЋР вЂљР В Р’В°Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљ"},
	RoleSupportSenior:             {Department: DepartmentSupport, Level: 2, Label: "Р В Р Р‹Р РЋРІР‚С™Р В Р’В°Р РЋР вЂљР РЋРІвЂљВ¬Р В РЎвЂР В РІвЂћвЂ“ Р РЋР С“Р В РЎвЂ”Р В Р’ВµР РЋРІР‚В Р В РЎвЂР В Р’В°Р В Р’В»Р В РЎвЂР РЋР С“Р РЋРІР‚С™ Р РЋРІР‚С™Р В Р’ВµР РЋРІР‚В¦Р В РЎвЂ”Р В РЎвЂўР В РўвЂР В РўвЂР В Р’ВµР РЋР вЂљР В Р’В¶Р В РЎвЂќР В РЎвЂ"},
	RoleSupportEmployee:           {Department: DepartmentSupport, Level: 1, Label: "Р В Р Р‹Р В РЎвЂ”Р В Р’ВµР РЋРІР‚В Р В РЎвЂР В Р’В°Р В Р’В»Р В РЎвЂР РЋР С“Р РЋРІР‚С™ Р РЋРІР‚С™Р В Р’ВµР РЋРІР‚В¦Р В РЎвЂ”Р В РЎвЂўР В РўвЂР В РўвЂР В Р’ВµР РЋР вЂљР В Р’В¶Р В РЎвЂќР В РЎвЂ"},
	RoleTechnicalHead:             {Department: DepartmentTechnical, Level: 3, Label: "Р В Р’В Р РЋРЎвЂњР В РЎвЂќР В РЎвЂўР В Р вЂ Р В РЎвЂўР В РўвЂР В РЎвЂР РЋРІР‚С™Р В Р’ВµР В Р’В»Р РЋР Р‰ Р РЋРІР‚С™Р В Р’ВµР РЋРІР‚В¦Р В Р вЂ¦Р В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР РЋРІР‚С™Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°"},
	RoleTechnicalSenior:           {Department: DepartmentTechnical, Level: 2, Label: "Р В Р Р‹Р РЋРІР‚С™Р В Р’В°Р РЋР вЂљР РЋРІвЂљВ¬Р В РЎвЂР В РІвЂћвЂ“ Р РЋРІР‚С™Р В Р’ВµР РЋРІР‚В¦Р В Р вЂ¦Р В РЎвЂР В РЎвЂќ"},
	RoleTechnicalEmployee:         {Department: DepartmentTechnical, Level: 1, Label: "Р В РЎС›Р В Р’ВµР РЋРІР‚В¦Р В Р вЂ¦Р В РЎвЂР В РЎвЂќ"},
	RoleTelecomDirector:           {Department: DepartmentTelecom, Level: 3, Label: "Р В РІР‚СњР В РЎвЂР РЋР вЂљР В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В РЎвЂўР РЋР вЂљ Р В РЎвЂ”Р В РЎвЂў Р РЋР С“Р РЋРІР‚С™Р РЋР вЂљР В РЎвЂўР В РЎвЂР РЋРІР‚С™Р В Р’ВµР В Р’В»Р РЋР Р‰Р РЋР С“Р РЋРІР‚С™Р В Р вЂ Р РЋРЎвЂњ"},
	RoleTelecomHead:               {Department: DepartmentTelecom, Level: 3, Label: "Р В Р’В Р РЋРЎвЂњР В РЎвЂќР В РЎвЂўР В Р вЂ Р В РЎвЂўР В РўвЂР В РЎвЂР РЋРІР‚С™Р В Р’ВµР В Р’В»Р РЋР Р‰ Р РЋР С“Р РЋРІР‚С™Р РЋР вЂљР В РЎвЂўР В РЎвЂР РЋРІР‚С™Р В Р’ВµР В Р’В»Р РЋР Р‰Р В Р вЂ¦Р В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР РЋРІР‚С™Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°"},
	RoleTelecomSeniorVOLS:         {Department: DepartmentTelecom, Level: 2, Label: "Р В Р Р‹Р РЋРІР‚С™Р В Р’В°Р РЋР вЂљР РЋРІвЂљВ¬Р В РЎвЂР В РІвЂћвЂ“ Р В РЎВР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р В Р’В°Р В Р’В¶Р В Р вЂ¦Р В РЎвЂР В РЎвЂќ Р В РІР‚в„ўР В РЎвЂєР В РІР‚С”Р В Р Р‹"},
	RoleTelecomSeniorLVS:          {Department: DepartmentTelecom, Level: 2, Label: "Р В Р Р‹Р РЋРІР‚С™Р В Р’В°Р РЋР вЂљР РЋРІвЂљВ¬Р В РЎвЂР В РІвЂћвЂ“ Р В РЎВР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р В Р’В°Р В Р’В¶Р В Р вЂ¦Р В РЎвЂР В РЎвЂќ Р В РІР‚С”Р В РІР‚в„ўР В Р Р‹"},
	RoleTelecomEmployeeVOLS:       {Department: DepartmentTelecom, Level: 1, Label: "Р В РЎС™Р В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р В Р’В°Р В Р’В¶Р В Р вЂ¦Р В РЎвЂР В РЎвЂќ Р В РІР‚в„ўР В РЎвЂєР В РІР‚С”Р В Р Р‹"},
	RoleTelecomEmployeeLVS:        {Department: DepartmentTelecom, Level: 1, Label: "Р В РЎС™Р В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р В Р’В°Р В Р’В¶Р В Р вЂ¦Р В РЎвЂР В РЎвЂќ Р В РІР‚С”Р В РІР‚в„ўР В Р Р‹"},
	RoleSKUDHead:                  {Department: DepartmentSKUD, Level: 3, Label: "Р В Р’В Р РЋРЎвЂњР В РЎвЂќР В РЎвЂўР В Р вЂ Р В РЎвЂўР В РўвЂР В РЎвЂР РЋРІР‚С™Р В Р’ВµР В Р’В»Р РЋР Р‰ Р В РЎвЂўР РЋРІР‚С™Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В° Р РЋРІР‚С™Р В Р’ВµР РЋРІР‚В¦Р В Р вЂ¦Р В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР С“Р В Р’В»Р РЋРЎвЂњР В Р’В¶Р В РЎвЂР В Р вЂ Р В Р’В°Р В Р вЂ¦Р В РЎвЂР РЋР РЏ Р В Р Р‹Р В РЎв„ўР В Р в‚¬Р В РІР‚Сњ"},
	RoleSKUDProjectManager:        {Department: DepartmentSKUD, Level: 2, Label: "Р В РЎС™Р В Р’ВµР В Р вЂ¦Р В Р’ВµР В РўвЂР В Р’В¶Р В Р’ВµР РЋР вЂљ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В РЎвЂўР В Р вЂ  Р В Р Р‹Р В РЎв„ўР В Р в‚¬Р В РІР‚Сњ"},
	RoleSKUDSeniorService:         {Department: DepartmentSKUD, Level: 2, Label: "Р В Р Р‹Р РЋРІР‚С™Р В Р’В°Р РЋР вЂљР РЋРІвЂљВ¬Р В РЎвЂР В РІвЂћвЂ“ Р РЋР С“Р В Р’ВµР РЋР вЂљР В Р вЂ Р В РЎвЂР РЋР С“Р В Р вЂ¦Р РЋРІР‚в„–Р В РІвЂћвЂ“ Р В РЎвЂР В Р вЂ¦Р В Р’В¶Р В Р’ВµР В Р вЂ¦Р В Р’ВµР РЋР вЂљ Р В Р Р‹Р В РЎв„ўР В Р в‚¬Р В РІР‚Сњ"},
	RoleSKUDSeniorInstaller:       {Department: DepartmentSKUD, Level: 2, Label: "Р В Р Р‹Р РЋРІР‚С™Р В Р’В°Р РЋР вЂљР РЋРІвЂљВ¬Р В РЎвЂР В РІвЂћвЂ“ Р В РЎВР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р В Р’В°Р В Р’В¶Р В Р вЂ¦Р В РЎвЂР В РЎвЂќ Р В Р Р‹Р В РЎв„ўР В Р в‚¬Р В РІР‚Сњ"},
	RoleSKUDServiceEngineer:       {Department: DepartmentSKUD, Level: 1, Label: "Р В Р Р‹Р В Р’ВµР РЋР вЂљР В Р вЂ Р В РЎвЂР РЋР С“Р В Р вЂ¦Р РЋРІР‚в„–Р В РІвЂћвЂ“ Р В РЎвЂР В Р вЂ¦Р В Р’В¶Р В Р’ВµР В Р вЂ¦Р В Р’ВµР РЋР вЂљ Р В Р Р‹Р В РЎв„ўР В Р в‚¬Р В РІР‚Сњ"},
	RoleSKUDInstaller:             {Department: DepartmentSKUD, Level: 1, Label: "Р В РЎС™Р В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р В Р’В°Р В Р’В¶Р В Р вЂ¦Р В РЎвЂР В РЎвЂќ Р В Р Р‹Р В РЎв„ўР В Р в‚¬Р В РІР‚Сњ"},
	RoleApprovalHead:              {Department: DepartmentApproval, Level: 3, Label: "Р В Р’В Р РЋРЎвЂњР В РЎвЂќР В РЎвЂўР В Р вЂ Р В РЎвЂўР В РўвЂР В РЎвЂР РЋРІР‚С™Р В Р’ВµР В Р’В»Р РЋР Р‰ Р РЋР С“Р В РЎвЂўР В РЎвЂ“Р В Р’В»Р В Р’В°Р РЋР С“Р В РЎвЂўР В Р вЂ Р В Р’В°Р В Р вЂ¦Р В РЎвЂР РЋР РЏ"},
	RoleApprovalSenior:            {Department: DepartmentApproval, Level: 2, Label: "Р В Р Р‹Р РЋРІР‚С™Р В Р’В°Р РЋР вЂљР РЋРІвЂљВ¬Р В РЎвЂР В РІвЂћвЂ“ Р В РЎВР В Р’ВµР В Р вЂ¦Р В Р’ВµР В РўвЂР В Р’В¶Р В Р’ВµР РЋР вЂљ Р РЋР С“Р В РЎвЂўР В РЎвЂ“Р В Р’В»Р В Р’В°Р РЋР С“Р В РЎвЂўР В Р вЂ Р В Р’В°Р В Р вЂ¦Р В РЎвЂР РЋР РЏ"},
	RoleApprovalEmployee:          {Department: DepartmentApproval, Level: 1, Label: "Р В РЎС™Р В Р’ВµР В Р вЂ¦Р В Р’ВµР В РўвЂР В Р’В¶Р В Р’ВµР РЋР вЂљ Р В РЎвЂ”Р В РЎвЂў Р РЋР С“Р В РЎвЂўР В РЎвЂ“Р В Р’В»Р В Р’В°Р РЋР С“Р В РЎвЂўР В Р вЂ Р В Р’В°Р В Р вЂ¦Р В РЎвЂР РЋР вЂ№"},
	RoleMarketingHead:             {Department: DepartmentMarketing, Level: 3, Label: "Р В Р’В Р РЋРЎвЂњР В РЎвЂќР В РЎвЂўР В Р вЂ Р В РЎвЂўР В РўвЂР В РЎвЂР РЋРІР‚С™Р В Р’ВµР В Р’В»Р РЋР Р‰ Р В РЎвЂўР РЋРІР‚С™Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В° Р РЋР вЂљР В Р’ВµР В РЎвЂќР В Р’В»Р В Р’В°Р В РЎВР РЋРІР‚в„– Р В РЎвЂ Р В РЎВР В Р’В°Р РЋР вЂљР В РЎвЂќР В Р’ВµР РЋРІР‚С™Р В РЎвЂР В Р вЂ¦Р В РЎвЂ“Р В Р’В°"},
	RoleMarketingCourier:          {Department: DepartmentMarketing, Level: 1, Label: "Р В РЎв„ўР РЋРЎвЂњР РЋР вЂљР РЋР Р‰Р В Р’ВµР РЋР вЂљ"},
	RoleCommercialDirector:        {Department: DepartmentCommercial, Level: 4, Label: "Р В Р’В Р РЋРЎвЂњР В РЎвЂќР В РЎвЂўР В Р вЂ Р В РЎвЂўР В РўвЂР В РЎвЂР РЋРІР‚С™Р В Р’ВµР В Р’В»Р РЋР Р‰ Р В РЎвЂўР РЋРІР‚С™Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В° Р РЋР вЂљР В Р’ВµР В РЎвЂќР В Р’В»Р В Р’В°Р В РЎВР РЋРІР‚в„– Р В РЎвЂ Р В РЎВР В Р’В°Р РЋР вЂљР В РЎвЂќР В Р’ВµР РЋРІР‚С™Р В РЎвЂР В Р вЂ¦Р В РЎвЂ“Р В Р’В°"},
	RoleCommercialSubscriberHead:  {Department: DepartmentCommercial, Level: 3, Label: "Р В Р’В Р РЋРЎвЂњР В РЎвЂќР В РЎвЂўР В Р вЂ Р В РЎвЂўР В РўвЂР В РЎвЂР РЋРІР‚С™Р В Р’ВµР В Р’В»Р РЋР Р‰ Р В Р’В°Р В Р’В±Р В РЎвЂўР В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР РЋРІР‚С™Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°"},
	RoleCommercialActiveSalesHead: {Department: DepartmentCommercial, Level: 3, Label: "Р В РЎС™Р В Р’ВµР В Р вЂ¦Р В Р’ВµР В РўвЂР В Р’В¶Р В Р’ВµР РЋР вЂљ Р В Р’В°Р В РЎвЂќР РЋРІР‚С™Р В РЎвЂР В Р вЂ Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В РўвЂР В Р’В°Р В Р’В¶"},
	RoleCommercialSeniorMRK:       {Department: DepartmentCommercial, Level: 2, Label: "Р В Р Р‹Р РЋРІР‚С™Р В Р’В°Р РЋР вЂљР РЋРІвЂљВ¬Р В РЎвЂР В РІвЂћвЂ“ Р В РЎС™Р В Р’В Р В РЎв„ў"},
	RoleCommercialSeniorMRYU:      {Department: DepartmentCommercial, Level: 2, Label: "Р В Р Р‹Р РЋРІР‚С™Р В Р’В°Р РЋР вЂљР РЋРІвЂљВ¬Р В РЎвЂР В РІвЂћвЂ“ Р В РЎС™Р В Р’В Р В Р’В®"},
	RoleCommercialEmployeeMRK:     {Department: DepartmentCommercial, Level: 1, Label: "Р В РЎС™Р В Р’В Р В РЎв„ў"},
	RoleCommercialEmployeeMRYU:    {Department: DepartmentCommercial, Level: 1, Label: "Р В РЎС™Р В Р’В Р В Р’В®"},
	RoleFinanceHead:               {Department: DepartmentFinance, Level: 3, Label: "Р В РІР‚СљР В Р’В». Р В Р’В±Р РЋРЎвЂњР РЋРІР‚В¦Р В РЎвЂ“Р В Р’В°Р В Р’В»Р РЋРІР‚С™Р В Р’ВµР РЋР вЂљ"},
	RoleFinanceEmployee:           {Department: DepartmentFinance, Level: 1, Label: "Р В РЎСџР В РЎвЂўР В РЎВР В РЎвЂўР РЋРІР‚В°Р В Р вЂ¦Р В РЎвЂР В РЎвЂќ Р В Р’В±Р РЋРЎвЂњР РЋРІР‚В¦Р В РЎвЂ“Р В Р’В°Р В Р’В»Р РЋРІР‚С™Р В Р’ВµР РЋР вЂљР В Р’В°"},
	RoleLegalEmployee:             {Department: DepartmentLegal, Level: 1, Label: "Р В Р’В®Р РЋР вЂљР В РЎвЂР РЋР С“Р РЋРІР‚С™"},
	RoleDevelopmentHead:           {Department: DepartmentDevelopment, Level: 3, Label: "Р В Р’В Р РЋРЎвЂњР В РЎвЂќР В РЎвЂўР В Р вЂ Р В РЎвЂўР В РўвЂР В РЎвЂР РЋРІР‚С™Р В Р’ВµР В Р’В»Р РЋР Р‰ Р В РЎвЂ“Р РЋР вЂљР РЋРЎвЂњР В РЎвЂ”Р В РЎвЂ”Р РЋРІР‚в„– Р РЋР вЂљР В Р’В°Р В Р’В·Р РЋР вЂљР В Р’В°Р В Р’В±Р В РЎвЂўР РЋРІР‚С™Р В РЎвЂќР В РЎвЂ"},
	RoleDevelopmentSenior:         {Department: DepartmentDevelopment, Level: 2, Label: "Р В Р Р‹Р РЋРІР‚С™Р В Р’В°Р РЋР вЂљР РЋРІвЂљВ¬Р В РЎвЂР В РІвЂћвЂ“ Р РЋР вЂљР В Р’В°Р В Р’В·Р РЋР вЂљР В Р’В°Р В Р’В±Р В РЎвЂўР РЋРІР‚С™Р РЋРІР‚РЋР В РЎвЂР В РЎвЂќ"},
	RoleDevelopmentEmployee:       {Department: DepartmentDevelopment, Level: 1, Label: "Р В Р’В Р В Р’В°Р В Р’В·Р РЋР вЂљР В Р’В°Р В Р’В±Р В РЎвЂўР РЋРІР‚С™Р РЋРІР‚РЋР В РЎвЂР В РЎвЂќ"},
}

var allBusinessDepartments = []string{
	DepartmentSupport,
	DepartmentTechnical,
	DepartmentTelecom,
	DepartmentSKUD,
	DepartmentApproval,
	DepartmentMarketing,
	DepartmentCommercial,
	DepartmentFinance,
	DepartmentLegal,
	DepartmentDevelopment,
}

// RU: Р В РЎС›Р В РЎвЂР В РЎвЂ” Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `Service`.
// EN: Data type `Service`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂўР В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р РЋР С“Р РЋРІР‚С™Р РЋР вЂљР РЋРЎвЂњР В РЎвЂќР РЋРІР‚С™Р РЋРЎвЂњР РЋР вЂљР РЋРЎвЂњ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `Service`, Р В РЎвЂќР В РЎвЂўР РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В Р’В°Р РЋР РЏ Р РЋРЎвЂњР РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р вЂ Р РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В Р вЂ  Р В Р’В±Р В РЎвЂР В Р’В·Р В Р вЂ¦Р В Р’ВµР РЋР С“-Р В Р’В»Р В РЎвЂўР В РЎвЂ“Р В РЎвЂР В РЎвЂќР В Р’Вµ, API Р В РЎвЂР В Р’В»Р В РЎвЂ Р РЋРІР‚С™Р В Р’ВµР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚В¦.
// EN: What it does: Service describes one billable position available to the current user: pricing, unit, ownership and allocation settings.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р В Р вЂ Р В Р’В°Р В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РЎвЂќР В Р’В°Р В РЎвЂќ Р В РЎвЂќР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р РЋР вЂљР В Р’В°Р В РЎвЂќР РЋРІР‚С™ Р В РЎвЂР В Р’В»Р В РЎвЂ Р В РЎвЂўР В РЎвЂ”Р В РЎвЂўР РЋР вЂљР В Р вЂ¦Р В Р’В°Р РЋР РЏ Р РЋРІР‚С™Р В РЎвЂўР РЋРІР‚РЋР В РЎвЂќР В Р’В° Р В РўвЂР В Р’В»Р РЋР РЏ Р В РўвЂР РЋР вЂљР РЋРЎвЂњР В РЎвЂ“Р В РЎвЂР РЋРІР‚В¦ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р’ВµР В РІвЂћвЂ“ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°; Р В РЎвЂР В Р’В·Р В РЎВР В Р’ВµР В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР РЋР РЏ Р В Р’В·Р В РўвЂР В Р’ВµР РЋР С“Р РЋР Р‰ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В РЎвЂў Р РЋРІР‚С™Р РЋР вЂљР В Р’ВµР В Р’В±Р РЋРЎвЂњР РЋР вЂ№Р РЋРІР‚С™ Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В РЎвЂўР В Р’В¶Р В Р вЂ¦Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂ.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type Service struct {
	ID                int64    `json:"id"`
	Code              string   `json:"code"`
	Name              string   `json:"name"`
	Unit              string   `json:"unit"`
	Rate              int      `json:"rate"`
	Description       string   `json:"description"`
	Category          string   `json:"category"`
	AllocationPercent *float64 `json:"allocationPercent,omitempty"`
	CreatedBy         string   `json:"createdBy,omitempty"`
	CreatedAt         string   `json:"createdAt,omitempty"`
}

// RU: Р В РЎС›Р В РЎвЂР В РЎвЂ” Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `UpsertServiceRequest`.
// EN: Data type `UpsertServiceRequest`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂўР В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р РЋР С“Р РЋРІР‚С™Р РЋР вЂљР РЋРЎвЂњР В РЎвЂќР РЋРІР‚С™Р РЋРЎвЂњР РЋР вЂљР РЋРЎвЂњ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `UpsertServiceRequest`, Р В РЎвЂќР В РЎвЂўР РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В Р’В°Р РЋР РЏ Р РЋРЎвЂњР РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р вЂ Р РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В Р вЂ  Р В Р’В±Р В РЎвЂР В Р’В·Р В Р вЂ¦Р В Р’ВµР РЋР С“-Р В Р’В»Р В РЎвЂўР В РЎвЂ“Р В РЎвЂР В РЎвЂќР В Р’Вµ, API Р В РЎвЂР В Р’В»Р В РЎвЂ Р РЋРІР‚С™Р В Р’ВµР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚В¦.
// EN: What it does: UpsertServiceRequest is the payload used by the frontend when a user creates or edits a service definition.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р В Р вЂ Р В Р’В°Р В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РЎвЂќР В Р’В°Р В РЎвЂќ Р В РЎвЂќР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р РЋР вЂљР В Р’В°Р В РЎвЂќР РЋРІР‚С™ Р В РЎвЂР В Р’В»Р В РЎвЂ Р В РЎвЂўР В РЎвЂ”Р В РЎвЂўР РЋР вЂљР В Р вЂ¦Р В Р’В°Р РЋР РЏ Р РЋРІР‚С™Р В РЎвЂўР РЋРІР‚РЋР В РЎвЂќР В Р’В° Р В РўвЂР В Р’В»Р РЋР РЏ Р В РўвЂР РЋР вЂљР РЋРЎвЂњР В РЎвЂ“Р В РЎвЂР РЋРІР‚В¦ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р’ВµР В РІвЂћвЂ“ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°; Р В РЎвЂР В Р’В·Р В РЎВР В Р’ВµР В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР РЋР РЏ Р В Р’В·Р В РўвЂР В Р’ВµР РЋР С“Р РЋР Р‰ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В РЎвЂў Р РЋРІР‚С™Р РЋР вЂљР В Р’ВµР В Р’В±Р РЋРЎвЂњР РЋР вЂ№Р РЋРІР‚С™ Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В РЎвЂўР В Р’В¶Р В Р вЂ¦Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂ.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type UpsertServiceRequest struct {
	ID                int64    `json:"id"`
	Name              string   `json:"name"`
	Unit              string   `json:"unit"`
	Rate              int      `json:"rate"`
	Category          string   `json:"category"`
	AllocationPercent *float64 `json:"allocationPercent"`
}

// RU: Р В РЎС›Р В РЎвЂР В РЎвЂ” Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `User`.
// EN: Data type `User`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂўР В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р РЋР С“Р РЋРІР‚С™Р РЋР вЂљР РЋРЎвЂњР В РЎвЂќР РЋРІР‚С™Р РЋРЎвЂњР РЋР вЂљР РЋРЎвЂњ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `User`, Р В РЎвЂќР В РЎвЂўР РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В Р’В°Р РЋР РЏ Р РЋРЎвЂњР РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р вЂ Р РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В Р вЂ  Р В Р’В±Р В РЎвЂР В Р’В·Р В Р вЂ¦Р В Р’ВµР РЋР С“-Р В Р’В»Р В РЎвЂўР В РЎвЂ“Р В РЎвЂР В РЎвЂќР В Р’Вµ, API Р В РЎвЂР В Р’В»Р В РЎвЂ Р РЋРІР‚С™Р В Р’ВµР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚В¦.
// EN: What it does: User stores public account information that can be safely returned to the frontend without password data.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р В Р вЂ Р В Р’В°Р В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РЎвЂќР В Р’В°Р В РЎвЂќ Р В РЎвЂќР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р РЋР вЂљР В Р’В°Р В РЎвЂќР РЋРІР‚С™ Р В РЎвЂР В Р’В»Р В РЎвЂ Р В РЎвЂўР В РЎвЂ”Р В РЎвЂўР РЋР вЂљР В Р вЂ¦Р В Р’В°Р РЋР РЏ Р РЋРІР‚С™Р В РЎвЂўР РЋРІР‚РЋР В РЎвЂќР В Р’В° Р В РўвЂР В Р’В»Р РЋР РЏ Р В РўвЂР РЋР вЂљР РЋРЎвЂњР В РЎвЂ“Р В РЎвЂР РЋРІР‚В¦ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р’ВµР В РІвЂћвЂ“ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°; Р В РЎвЂР В Р’В·Р В РЎВР В Р’ВµР В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР РЋР РЏ Р В Р’В·Р В РўвЂР В Р’ВµР РЋР С“Р РЋР Р‰ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В РЎвЂў Р РЋРІР‚С™Р РЋР вЂљР В Р’ВµР В Р’В±Р РЋРЎвЂњР РЋР вЂ№Р РЋРІР‚С™ Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В РЎвЂўР В Р’В¶Р В Р вЂ¦Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂ.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type User struct {
	ID                    int64  `json:"id"`
	Username              string `json:"username"`
	FullName              string `json:"fullName"`
	Role                  string `json:"role"`
	CreatedAt             string `json:"createdAt"`
	LastActNumber         int    `json:"lastActNumber"`
	PreferredContractCode string `json:"preferredContractCode"`
	ContractSPBKSNumber   string `json:"contractSPBKSNumber"`
	ContractGrizablNumber string `json:"contractGrizablNumber"`
	ContractSignedAt      string `json:"contractSignedAt"`
}

// RU: Р В РЎС›Р В РЎвЂР В РЎвЂ” Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `UserWithPassword`.
// EN: Data type `UserWithPassword`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂўР В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р РЋР С“Р РЋРІР‚С™Р РЋР вЂљР РЋРЎвЂњР В РЎвЂќР РЋРІР‚С™Р РЋРЎвЂњР РЋР вЂљР РЋРЎвЂњ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `UserWithPassword`, Р В РЎвЂќР В РЎвЂўР РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В Р’В°Р РЋР РЏ Р РЋРЎвЂњР РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р вЂ Р РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В Р вЂ  Р В Р’В±Р В РЎвЂР В Р’В·Р В Р вЂ¦Р В Р’ВµР РЋР С“-Р В Р’В»Р В РЎвЂўР В РЎвЂ“Р В РЎвЂР В РЎвЂќР В Р’Вµ, API Р В РЎвЂР В Р’В»Р В РЎвЂ Р РЋРІР‚С™Р В Р’ВµР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚В¦.
// EN: What it does: UserWithPassword extends a user payload with a plain-text password for creation workflows only.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р В Р вЂ Р В Р’В°Р В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РЎвЂќР В Р’В°Р В РЎвЂќ Р В РЎвЂќР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р РЋР вЂљР В Р’В°Р В РЎвЂќР РЋРІР‚С™ Р В РЎвЂР В Р’В»Р В РЎвЂ Р В РЎвЂўР В РЎвЂ”Р В РЎвЂўР РЋР вЂљР В Р вЂ¦Р В Р’В°Р РЋР РЏ Р РЋРІР‚С™Р В РЎвЂўР РЋРІР‚РЋР В РЎвЂќР В Р’В° Р В РўвЂР В Р’В»Р РЋР РЏ Р В РўвЂР РЋР вЂљР РЋРЎвЂњР В РЎвЂ“Р В РЎвЂР РЋРІР‚В¦ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р’ВµР В РІвЂћвЂ“ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°; Р В РЎвЂР В Р’В·Р В РЎВР В Р’ВµР В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР РЋР РЏ Р В Р’В·Р В РўвЂР В Р’ВµР РЋР С“Р РЋР Р‰ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В РЎвЂў Р РЋРІР‚С™Р РЋР вЂљР В Р’ВµР В Р’В±Р РЋРЎвЂњР РЋР вЂ№Р РЋРІР‚С™ Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В РЎвЂўР В Р’В¶Р В Р вЂ¦Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂ.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type UserWithPassword struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type ResetUserPasswordRequest struct {
	UserID   int64  `json:"userID"`
	Password string `json:"password"`
}

type UpdateUserFullNameRequest struct {
	UserID   int64  `json:"userID"`
	FullName string `json:"fullName"`
}

type UpdateUserContractDetailsRequest struct {
	UserID                int64  `json:"userID"`
	FullName              string `json:"fullName"`
	PreferredContractCode string `json:"preferredContractCode"`
	ContractSPBKSNumber   string `json:"contractSPBKSNumber"`
	ContractGrizablNumber string `json:"contractGrizablNumber"`
	ContractSignedAt      string `json:"contractSignedAt"`
}

type UpdateUserLastActNumberRequest struct {
	UserID        int64 `json:"userID"`
	LastActNumber int   `json:"lastActNumber"`
}

// RU: Р В РЎС›Р В РЎвЂР В РЎвЂ” Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `LoginRequest`.
// EN: Data type `LoginRequest`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂўР В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р РЋР С“Р РЋРІР‚С™Р РЋР вЂљР РЋРЎвЂњР В РЎвЂќР РЋРІР‚С™Р РЋРЎвЂњР РЋР вЂљР РЋРЎвЂњ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `LoginRequest`, Р В РЎвЂќР В РЎвЂўР РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В Р’В°Р РЋР РЏ Р РЋРЎвЂњР РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р вЂ Р РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В Р вЂ  Р В Р’В±Р В РЎвЂР В Р’В·Р В Р вЂ¦Р В Р’ВµР РЋР С“-Р В Р’В»Р В РЎвЂўР В РЎвЂ“Р В РЎвЂР В РЎвЂќР В Р’Вµ, API Р В РЎвЂР В Р’В»Р В РЎвЂ Р РЋРІР‚С™Р В Р’ВµР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚В¦.
// EN: What it does: LoginRequest carries credentials from the login form to the backend session logic.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р В Р вЂ Р В Р’В°Р В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РЎвЂќР В Р’В°Р В РЎвЂќ Р В РЎвЂќР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р РЋР вЂљР В Р’В°Р В РЎвЂќР РЋРІР‚С™ Р В РЎвЂР В Р’В»Р В РЎвЂ Р В РЎвЂўР В РЎвЂ”Р В РЎвЂўР РЋР вЂљР В Р вЂ¦Р В Р’В°Р РЋР РЏ Р РЋРІР‚С™Р В РЎвЂўР РЋРІР‚РЋР В РЎвЂќР В Р’В° Р В РўвЂР В Р’В»Р РЋР РЏ Р В РўвЂР РЋР вЂљР РЋРЎвЂњР В РЎвЂ“Р В РЎвЂР РЋРІР‚В¦ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р’ВµР В РІвЂћвЂ“ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°; Р В РЎвЂР В Р’В·Р В РЎВР В Р’ВµР В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР РЋР РЏ Р В Р’В·Р В РўвЂР В Р’ВµР РЋР С“Р РЋР Р‰ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В РЎвЂў Р РЋРІР‚С™Р РЋР вЂљР В Р’ВµР В Р’В±Р РЋРЎвЂњР РЋР вЂ№Р РЋРІР‚С™ Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В РЎвЂўР В Р’В¶Р В Р вЂ¦Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂ.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RU: Р В РЎС›Р В РЎвЂР В РЎвЂ” Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `SessionState`.
// EN: Data type `SessionState`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂўР В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р РЋР С“Р РЋРІР‚С™Р РЋР вЂљР РЋРЎвЂњР В РЎвЂќР РЋРІР‚С™Р РЋРЎвЂњР РЋР вЂљР РЋРЎвЂњ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `SessionState`, Р В РЎвЂќР В РЎвЂўР РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В Р’В°Р РЋР РЏ Р РЋРЎвЂњР РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р вЂ Р РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В Р вЂ  Р В Р’В±Р В РЎвЂР В Р’В·Р В Р вЂ¦Р В Р’ВµР РЋР С“-Р В Р’В»Р В РЎвЂўР В РЎвЂ“Р В РЎвЂР В РЎвЂќР В Р’Вµ, API Р В РЎвЂР В Р’В»Р В РЎвЂ Р РЋРІР‚С™Р В Р’ВµР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚В¦.
// EN: What it does: SessionState is the frontend-facing snapshot of the current authenticated user and their capabilities.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р В Р вЂ Р В Р’В°Р В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РЎвЂќР В Р’В°Р В РЎвЂќ Р В РЎвЂќР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р РЋР вЂљР В Р’В°Р В РЎвЂќР РЋРІР‚С™ Р В РЎвЂР В Р’В»Р В РЎвЂ Р В РЎвЂўР В РЎвЂ”Р В РЎвЂўР РЋР вЂљР В Р вЂ¦Р В Р’В°Р РЋР РЏ Р РЋРІР‚С™Р В РЎвЂўР РЋРІР‚РЋР В РЎвЂќР В Р’В° Р В РўвЂР В Р’В»Р РЋР РЏ Р В РўвЂР РЋР вЂљР РЋРЎвЂњР В РЎвЂ“Р В РЎвЂР РЋРІР‚В¦ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р’ВµР В РІвЂћвЂ“ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°; Р В РЎвЂР В Р’В·Р В РЎВР В Р’ВµР В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР РЋР РЏ Р В Р’В·Р В РўвЂР В Р’ВµР РЋР С“Р РЋР Р‰ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В РЎвЂў Р РЋРІР‚С™Р РЋР вЂљР В Р’ВµР В Р’В±Р РЋРЎвЂњР РЋР вЂ№Р РЋРІР‚С™ Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В РЎвЂўР В Р’В¶Р В Р вЂ¦Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂ.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type SessionState struct {
	Authenticated bool   `json:"authenticated"`
	User          *User  `json:"user,omitempty"`
	CanManage     bool   `json:"canManage"`
	CanAdmin      bool   `json:"canAdmin"`
	CanModerate   bool   `json:"canModerate"`
	Message       string `json:"message,omitempty"`
}

// RU: Р В РЎС›Р В РЎвЂР В РЎвЂ” Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `CalculationRequest`.
// EN: Data type `CalculationRequest`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂўР В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р РЋР С“Р РЋРІР‚С™Р РЋР вЂљР РЋРЎвЂњР В РЎвЂќР РЋРІР‚С™Р РЋРЎвЂњР РЋР вЂљР РЋРЎвЂњ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `CalculationRequest`, Р В РЎвЂќР В РЎвЂўР РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В Р’В°Р РЋР РЏ Р РЋРЎвЂњР РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р вЂ Р РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В Р вЂ  Р В Р’В±Р В РЎвЂР В Р’В·Р В Р вЂ¦Р В Р’ВµР РЋР С“-Р В Р’В»Р В РЎвЂўР В РЎвЂ“Р В РЎвЂР В РЎвЂќР В Р’Вµ, API Р В РЎвЂР В Р’В»Р В РЎвЂ Р РЋРІР‚С™Р В Р’ВµР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚В¦.
// EN: What it does: CalculationRequest contains the target amount and per-service weights used to build a calculation.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р В Р вЂ Р В Р’В°Р В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РЎвЂќР В Р’В°Р В РЎвЂќ Р В РЎвЂќР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р РЋР вЂљР В Р’В°Р В РЎвЂќР РЋРІР‚С™ Р В РЎвЂР В Р’В»Р В РЎвЂ Р В РЎвЂўР В РЎвЂ”Р В РЎвЂўР РЋР вЂљР В Р вЂ¦Р В Р’В°Р РЋР РЏ Р РЋРІР‚С™Р В РЎвЂўР РЋРІР‚РЋР В РЎвЂќР В Р’В° Р В РўвЂР В Р’В»Р РЋР РЏ Р В РўвЂР РЋР вЂљР РЋРЎвЂњР В РЎвЂ“Р В РЎвЂР РЋРІР‚В¦ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р’ВµР В РІвЂћвЂ“ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°; Р В РЎвЂР В Р’В·Р В РЎВР В Р’ВµР В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР РЋР РЏ Р В Р’В·Р В РўвЂР В Р’ВµР РЋР С“Р РЋР Р‰ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В РЎвЂў Р РЋРІР‚С™Р РЋР вЂљР В Р’ВµР В Р’В±Р РЋРЎвЂњР РЋР вЂ№Р РЋРІР‚С™ Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В РЎвЂўР В Р’В¶Р В Р вЂ¦Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂ.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type CalculationRequest struct {
	TargetAmount int            `json:"targetAmount"`
	Weights      map[string]int `json:"weights"`
}

// RU: Р В РЎС›Р В РЎвЂР В РЎвЂ” Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `CalculationItem`.
// EN: Data type `CalculationItem`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂўР В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р РЋР С“Р РЋРІР‚С™Р РЋР вЂљР РЋРЎвЂњР В РЎвЂќР РЋРІР‚С™Р РЋРЎвЂњР РЋР вЂљР РЋРЎвЂњ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `CalculationItem`, Р В РЎвЂќР В РЎвЂўР РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В Р’В°Р РЋР РЏ Р РЋРЎвЂњР РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р вЂ Р РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В Р вЂ  Р В Р’В±Р В РЎвЂР В Р’В·Р В Р вЂ¦Р В Р’ВµР РЋР С“-Р В Р’В»Р В РЎвЂўР В РЎвЂ“Р В РЎвЂР В РЎвЂќР В Р’Вµ, API Р В РЎвЂР В Р’В»Р В РЎвЂ Р РЋРІР‚С™Р В Р’ВµР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚В¦.
// EN: What it does: CalculationItem is one line of the generated or archived calculation with both business and UI-oriented fields.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р В Р вЂ Р В Р’В°Р В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РЎвЂќР В Р’В°Р В РЎвЂќ Р В РЎвЂќР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р РЋР вЂљР В Р’В°Р В РЎвЂќР РЋРІР‚С™ Р В РЎвЂР В Р’В»Р В РЎвЂ Р В РЎвЂўР В РЎвЂ”Р В РЎвЂўР РЋР вЂљР В Р вЂ¦Р В Р’В°Р РЋР РЏ Р РЋРІР‚С™Р В РЎвЂўР РЋРІР‚РЋР В РЎвЂќР В Р’В° Р В РўвЂР В Р’В»Р РЋР РЏ Р В РўвЂР РЋР вЂљР РЋРЎвЂњР В РЎвЂ“Р В РЎвЂР РЋРІР‚В¦ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р’ВµР В РІвЂћвЂ“ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°; Р В РЎвЂР В Р’В·Р В РЎВР В Р’ВµР В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР РЋР РЏ Р В Р’В·Р В РўвЂР В Р’ВµР РЋР С“Р РЋР Р‰ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В РЎвЂў Р РЋРІР‚С™Р РЋР вЂљР В Р’ВµР В Р’В±Р РЋРЎвЂњР РЋР вЂ№Р РЋРІР‚С™ Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В РЎвЂўР В Р’В¶Р В Р вЂ¦Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂ.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type CalculationItem struct {
	ServiceID         int64    `json:"serviceId"`
	ServiceCode       string   `json:"serviceCode"`
	Name              string   `json:"name"`
	Unit              string   `json:"unit"`
	Rate              int      `json:"rate"`
	Quantity          int      `json:"quantity"`
	LineTotal         int      `json:"lineTotal"`
	Description       string   `json:"description"`
	Weight            int      `json:"weight"`
	Category          string   `json:"category"`
	AllocationPercent *float64 `json:"allocationPercent,omitempty"`
}

// RU: Р В РЎС›Р В РЎвЂР В РЎвЂ” Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `CalculationResult`.
// EN: Data type `CalculationResult`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂўР В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р РЋР С“Р РЋРІР‚С™Р РЋР вЂљР РЋРЎвЂњР В РЎвЂќР РЋРІР‚С™Р РЋРЎвЂњР РЋР вЂљР РЋРЎвЂњ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `CalculationResult`, Р В РЎвЂќР В РЎвЂўР РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В Р’В°Р РЋР РЏ Р РЋРЎвЂњР РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р вЂ Р РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В Р вЂ  Р В Р’В±Р В РЎвЂР В Р’В·Р В Р вЂ¦Р В Р’ВµР РЋР С“-Р В Р’В»Р В РЎвЂўР В РЎвЂ“Р В РЎвЂР В РЎвЂќР В Р’Вµ, API Р В РЎвЂР В Р’В»Р В РЎвЂ Р РЋРІР‚С™Р В Р’ВµР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚В¦.
// EN: What it does: CalculationResult is the full response returned after a calculation attempt, including totals and metadata.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р В Р вЂ Р В Р’В°Р В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РЎвЂќР В Р’В°Р В РЎвЂќ Р В РЎвЂќР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р РЋР вЂљР В Р’В°Р В РЎвЂќР РЋРІР‚С™ Р В РЎвЂР В Р’В»Р В РЎвЂ Р В РЎвЂўР В РЎвЂ”Р В РЎвЂўР РЋР вЂљР В Р вЂ¦Р В Р’В°Р РЋР РЏ Р РЋРІР‚С™Р В РЎвЂўР РЋРІР‚РЋР В РЎвЂќР В Р’В° Р В РўвЂР В Р’В»Р РЋР РЏ Р В РўвЂР РЋР вЂљР РЋРЎвЂњР В РЎвЂ“Р В РЎвЂР РЋРІР‚В¦ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р’ВµР В РІвЂћвЂ“ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°; Р В РЎвЂР В Р’В·Р В РЎВР В Р’ВµР В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР РЋР РЏ Р В Р’В·Р В РўвЂР В Р’ВµР РЋР С“Р РЋР Р‰ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В РЎвЂў Р РЋРІР‚С™Р РЋР вЂљР В Р’ВµР В Р’В±Р РЋРЎвЂњР РЋР вЂ№Р РЋРІР‚С™ Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В РЎвЂўР В Р’В¶Р В Р вЂ¦Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂ.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type CalculationResult struct {
	TargetAmount   int               `json:"targetAmount"`
	TotalAmount    int               `json:"totalAmount"`
	Items          []CalculationItem `json:"items"`
	FoundExact     bool              `json:"foundExact"`
	GeneratedAt    string            `json:"generatedAt"`
	ActiveServices int               `json:"activeServices"`
	Weights        map[string]int    `json:"weights"`
}

// RU: Р В РЎС›Р В РЎвЂР В РЎвЂ” Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `SaveCalculationRequest`.
// EN: Data type `SaveCalculationRequest`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂўР В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р РЋР С“Р РЋРІР‚С™Р РЋР вЂљР РЋРЎвЂњР В РЎвЂќР РЋРІР‚С™Р РЋРЎвЂњР РЋР вЂљР РЋРЎвЂњ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `SaveCalculationRequest`, Р В РЎвЂќР В РЎвЂўР РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В Р’В°Р РЋР РЏ Р РЋРЎвЂњР РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р вЂ Р РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В Р вЂ  Р В Р’В±Р В РЎвЂР В Р’В·Р В Р вЂ¦Р В Р’ВµР РЋР С“-Р В Р’В»Р В РЎвЂўР В РЎвЂ“Р В РЎвЂР В РЎвЂќР В Р’Вµ, API Р В РЎвЂР В Р’В»Р В РЎвЂ Р РЋРІР‚С™Р В Р’ВµР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚В¦.
// EN: What it does: SaveCalculationRequest is sent when the frontend persists the current calculation into the archive.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р В Р вЂ Р В Р’В°Р В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РЎвЂќР В Р’В°Р В РЎвЂќ Р В РЎвЂќР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р РЋР вЂљР В Р’В°Р В РЎвЂќР РЋРІР‚С™ Р В РЎвЂР В Р’В»Р В РЎвЂ Р В РЎвЂўР В РЎвЂ”Р В РЎвЂўР РЋР вЂљР В Р вЂ¦Р В Р’В°Р РЋР РЏ Р РЋРІР‚С™Р В РЎвЂўР РЋРІР‚РЋР В РЎвЂќР В Р’В° Р В РўвЂР В Р’В»Р РЋР РЏ Р В РўвЂР РЋР вЂљР РЋРЎвЂњР В РЎвЂ“Р В РЎвЂР РЋРІР‚В¦ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р’ВµР В РІвЂћвЂ“ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°; Р В РЎвЂР В Р’В·Р В РЎВР В Р’ВµР В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР РЋР РЏ Р В Р’В·Р В РўвЂР В Р’ВµР РЋР С“Р РЋР Р‰ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В РЎвЂў Р РЋРІР‚С™Р РЋР вЂљР В Р’ВµР В Р’В±Р РЋРЎвЂњР РЋР вЂ№Р РЋРІР‚С™ Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В РЎвЂўР В Р’В¶Р В Р вЂ¦Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂ.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type SaveCalculationRequest struct {
	Title        string            `json:"title"`
	TargetAmount int               `json:"targetAmount"`
	Items        []CalculationItem `json:"items"`
}

// RU: Р В РЎС›Р В РЎвЂР В РЎвЂ” Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `SavedCalculation`.
// EN: Data type `SavedCalculation`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂўР В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р РЋР С“Р РЋРІР‚С™Р РЋР вЂљР РЋРЎвЂњР В РЎвЂќР РЋРІР‚С™Р РЋРЎвЂњР РЋР вЂљР РЋРЎвЂњ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `SavedCalculation`, Р В РЎвЂќР В РЎвЂўР РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В Р’В°Р РЋР РЏ Р РЋРЎвЂњР РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р вЂ Р РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В Р вЂ  Р В Р’В±Р В РЎвЂР В Р’В·Р В Р вЂ¦Р В Р’ВµР РЋР С“-Р В Р’В»Р В РЎвЂўР В РЎвЂ“Р В РЎвЂР В РЎвЂќР В Р’Вµ, API Р В РЎвЂР В Р’В»Р В РЎвЂ Р РЋРІР‚С™Р В Р’ВµР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚В¦.
// EN: What it does: SavedCalculation represents one archived calculation entry as stored in SQLite and shown in history.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р В Р вЂ Р В Р’В°Р В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РЎвЂќР В Р’В°Р В РЎвЂќ Р В РЎвЂќР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р РЋР вЂљР В Р’В°Р В РЎвЂќР РЋРІР‚С™ Р В РЎвЂР В Р’В»Р В РЎвЂ Р В РЎвЂўР В РЎвЂ”Р В РЎвЂўР РЋР вЂљР В Р вЂ¦Р В Р’В°Р РЋР РЏ Р РЋРІР‚С™Р В РЎвЂўР РЋРІР‚РЋР В РЎвЂќР В Р’В° Р В РўвЂР В Р’В»Р РЋР РЏ Р В РўвЂР РЋР вЂљР РЋРЎвЂњР В РЎвЂ“Р В РЎвЂР РЋРІР‚В¦ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р’ВµР В РІвЂћвЂ“ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°; Р В РЎвЂР В Р’В·Р В РЎВР В Р’ВµР В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР РЋР РЏ Р В Р’В·Р В РўвЂР В Р’ВµР РЋР С“Р РЋР Р‰ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В РЎвЂў Р РЋРІР‚С™Р РЋР вЂљР В Р’ВµР В Р’В±Р РЋРЎвЂњР РЋР вЂ№Р РЋРІР‚С™ Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В РЎвЂўР В Р’В¶Р В Р вЂ¦Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂ.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type SavedCalculation struct {
	ID           int64             `json:"id"`
	Title        string            `json:"title"`
	TargetAmount int               `json:"targetAmount"`
	TotalAmount  int               `json:"totalAmount"`
	Items        []CalculationItem `json:"items"`
	CreatedAt    string            `json:"createdAt"`
	CreatedBy    string            `json:"createdBy"`
	CreatedRole  string            `json:"createdRole,omitempty"`
}

type ExportCalculationRequest struct {
	ActNumber        int               `json:"actNumber"`
	EmployeeFullName string            `json:"employeeFullName"`
	ContractCode     string            `json:"contractCode"`
	ContractNumber   string            `json:"contractNumber"`
	ContractDate     string            `json:"contractDate"`
	GeneratedAt      string            `json:"generatedAt"`
	TargetAmount     int               `json:"targetAmount"`
	Items            []CalculationItem `json:"items"`
}

// RU: Р В РЎС›Р В РЎвЂР В РЎвЂ” Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `CopyArchiveServicesResult`.
// EN: Data type `CopyArchiveServicesResult`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂўР В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р РЋР вЂљР В Р’ВµР В Р’В·Р РЋРЎвЂњР В Р’В»Р РЋР Р‰Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚С™ Р В РЎвЂќР В РЎвЂўР В РЎвЂ”Р В РЎвЂР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’В°Р В Р вЂ¦Р В РЎвЂР РЋР РЏ Р РЋРЎвЂњР РЋР С“Р В Р’В»Р РЋРЎвЂњР В РЎвЂ“ Р В РЎвЂР В Р’В· Р В Р’В°Р РЋР вЂљР РЋРІР‚В¦Р В РЎвЂР В Р вЂ Р В Р вЂ¦Р В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р РЋР вЂљР В Р’В°Р РЋР С“Р РЋРІР‚РЋР РЋРІР‚ВР РЋРІР‚С™Р В Р’В° Р В Р вЂ  Р РЋР С“Р В РЎвЂ”Р В РЎвЂР РЋР С“Р В РЎвЂўР В РЎвЂќ Р РЋРЎвЂњР РЋР С“Р В Р’В»Р РЋРЎвЂњР В РЎвЂ“ Р В Р’В°Р В РўвЂР В РЎВР В РЎвЂР В Р вЂ¦Р В РЎвЂР РЋР С“Р РЋРІР‚С™Р РЋР вЂљР В Р’В°Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В Р’В°.
// EN: What it does: CopyArchiveServicesResult reports how many services were created or updated after importing from an archived calculation.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р В РЎвЂР РЋР С“Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р РЋР Р‰Р В Р’В·Р РЋРЎвЂњР В Р’ВµР РЋРІР‚С™Р РЋР С“Р РЋР РЏ Р РЋРІР‚С™Р В РЎвЂўР В Р’В»Р РЋР Р‰Р В РЎвЂќР В РЎвЂў Р В Р вЂ  Р В Р’В°Р В РўвЂР В РЎВР В РЎвЂР В Р вЂ¦Р РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎВ Р РЋР С“Р РЋРІР‚В Р В Р’ВµР В Р вЂ¦Р В Р’В°Р РЋР вЂљР В РЎвЂР В РЎвЂ; Р В РЎвЂ”Р В РЎвЂўР В РЎВР В РЎвЂўР В РЎвЂ“Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р РЋРІР‚С›Р РЋР вЂљР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р В Р’ВµР В Р вЂ¦Р В РўвЂР РЋРЎвЂњ Р В РЎвЂ”Р В РЎвЂўР В РЎвЂќР В Р’В°Р В Р’В·Р В Р’В°Р РЋРІР‚С™Р РЋР Р‰ Р В РЎвЂ”Р В РЎвЂўР В Р вЂ¦Р РЋР РЏР РЋРІР‚С™Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р РЋР С“Р В РЎвЂўР В РЎвЂўР В Р’В±Р РЋРІР‚В°Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР РЋР С“Р В Р’В»Р В Р’Вµ Р В РЎвЂР В РЎВР В РЎвЂ”Р В РЎвЂўР РЋР вЂљР РЋРІР‚С™Р В Р’В°; Р В РЎвЂўР РЋРІР‚С™Р В РўвЂР В Р’ВµР В Р’В»Р РЋР РЏР В Р’ВµР РЋРІР‚С™ Р РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚С™Р В РЎвЂР РЋР С“Р РЋРІР‚С™Р В РЎвЂР В РЎвЂќР РЋРЎвЂњ Р В РЎвЂўР В РЎвЂ”Р В Р’ВµР РЋР вЂљР В Р’В°Р РЋРІР‚В Р В РЎвЂР В РЎвЂ Р В РЎвЂўР РЋРІР‚С™ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В Р вЂ¦Р В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р РЋР С“Р В РЎвЂ”Р В РЎвЂР РЋР С“Р В РЎвЂќР В Р’В° Р РЋРЎвЂњР РЋР С“Р В Р’В»Р РЋРЎвЂњР В РЎвЂ“.
// EN: Key points: used only in the admin-only archive import flow; lets the frontend show a concise status message; keeps operation stats separate from the full services list.
type CopyArchiveServicesResult struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
}

// RU: Р В РЎС›Р В РЎвЂР В РЎвЂ” Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `AppBootstrap`.
// EN: Data type `AppBootstrap`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂўР В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р РЋР С“Р РЋРІР‚С™Р РЋР вЂљР РЋРЎвЂњР В РЎвЂќР РЋРІР‚С™Р РЋРЎвЂњР РЋР вЂљР РЋРЎвЂњ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `AppBootstrap`, Р В РЎвЂќР В РЎвЂўР РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В Р’В°Р РЋР РЏ Р РЋРЎвЂњР РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р вЂ Р РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В Р вЂ  Р В Р’В±Р В РЎвЂР В Р’В·Р В Р вЂ¦Р В Р’ВµР РЋР С“-Р В Р’В»Р В РЎвЂўР В РЎвЂ“Р В РЎвЂР В РЎвЂќР В Р’Вµ, API Р В РЎвЂР В Р’В»Р В РЎвЂ Р РЋРІР‚С™Р В Р’ВµР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚В¦.
// EN: What it does: AppBootstrap aggregates all initial data the frontend needs after startup or refresh.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р В Р вЂ Р В Р’В°Р В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РЎвЂќР В Р’В°Р В РЎвЂќ Р В РЎвЂќР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р РЋР вЂљР В Р’В°Р В РЎвЂќР РЋРІР‚С™ Р В РЎвЂР В Р’В»Р В РЎвЂ Р В РЎвЂўР В РЎвЂ”Р В РЎвЂўР РЋР вЂљР В Р вЂ¦Р В Р’В°Р РЋР РЏ Р РЋРІР‚С™Р В РЎвЂўР РЋРІР‚РЋР В РЎвЂќР В Р’В° Р В РўвЂР В Р’В»Р РЋР РЏ Р В РўвЂР РЋР вЂљР РЋРЎвЂњР В РЎвЂ“Р В РЎвЂР РЋРІР‚В¦ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р’ВµР В РІвЂћвЂ“ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°; Р В РЎвЂР В Р’В·Р В РЎВР В Р’ВµР В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР РЋР РЏ Р В Р’В·Р В РўвЂР В Р’ВµР РЋР С“Р РЋР Р‰ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В РЎвЂў Р РЋРІР‚С™Р РЋР вЂљР В Р’ВµР В Р’В±Р РЋРЎвЂњР РЋР вЂ№Р РЋРІР‚С™ Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В РЎвЂўР В Р’В¶Р В Р вЂ¦Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂ.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type AppBootstrap struct {
	Session             SessionState       `json:"session"`
	Services            []Service          `json:"services"`
	Users               []User             `json:"users"`
	SavedCalculations   []SavedCalculation `json:"savedCalculations"`
	DefaultGroupPercent map[string]float64 `json:"defaultGroupPercent"`
}

// RU: Р В РЎС›Р В РЎвЂР В РЎвЂ” Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `allocationState`.
// EN: Data type `allocationState`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂўР В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р РЋР С“Р РЋРІР‚С™Р РЋР вЂљР РЋРЎвЂњР В РЎвЂќР РЋРІР‚С™Р РЋРЎвЂњР РЋР вЂљР РЋРЎвЂњ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `allocationState`, Р В РЎвЂќР В РЎвЂўР РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В Р’В°Р РЋР РЏ Р РЋРЎвЂњР РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р вЂ Р РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В Р вЂ  Р В Р’В±Р В РЎвЂР В Р’В·Р В Р вЂ¦Р В Р’ВµР РЋР С“-Р В Р’В»Р В РЎвЂўР В РЎвЂ“Р В РЎвЂР В РЎвЂќР В Р’Вµ, API Р В РЎвЂР В Р’В»Р В РЎвЂ Р РЋРІР‚С™Р В Р’ВµР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚В¦.
// EN: What it does: allocationState is an internal DP cell used while searching for a good quantity distribution.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р В Р вЂ Р В Р’В°Р В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РЎвЂќР В Р’В°Р В РЎвЂќ Р В РЎвЂќР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р РЋР вЂљР В Р’В°Р В РЎвЂќР РЋРІР‚С™ Р В РЎвЂР В Р’В»Р В РЎвЂ Р В РЎвЂўР В РЎвЂ”Р В РЎвЂўР РЋР вЂљР В Р вЂ¦Р В Р’В°Р РЋР РЏ Р РЋРІР‚С™Р В РЎвЂўР РЋРІР‚РЋР В РЎвЂќР В Р’В° Р В РўвЂР В Р’В»Р РЋР РЏ Р В РўвЂР РЋР вЂљР РЋРЎвЂњР В РЎвЂ“Р В РЎвЂР РЋРІР‚В¦ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р’ВµР В РІвЂћвЂ“ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°; Р В РЎвЂР В Р’В·Р В РЎВР В Р’ВµР В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР РЋР РЏ Р В Р’В·Р В РўвЂР В Р’ВµР РЋР С“Р РЋР Р‰ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В РЎвЂў Р РЋРІР‚С™Р РЋР вЂљР В Р’ВµР В Р’В±Р РЋРЎвЂњР РЋР вЂ№Р РЋРІР‚С™ Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В РЎвЂўР В Р’В¶Р В Р вЂ¦Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂ.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type allocationState struct {
	distance int
	score    int
	count    int
	prev     int
	idx      int
	ok       bool
}

// RU: Р В РЎС›Р В РЎвЂР В РЎвЂ” Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `groupAllocation`.
// EN: Data type `groupAllocation`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂўР В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р РЋР С“Р РЋРІР‚С™Р РЋР вЂљР РЋРЎвЂњР В РЎвЂќР РЋРІР‚С™Р РЋРЎвЂњР РЋР вЂљР РЋРЎвЂњ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `groupAllocation`, Р В РЎвЂќР В РЎвЂўР РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В Р’В°Р РЋР РЏ Р РЋРЎвЂњР РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р вЂ Р РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В Р вЂ  Р В Р’В±Р В РЎвЂР В Р’В·Р В Р вЂ¦Р В Р’ВµР РЋР С“-Р В Р’В»Р В РЎвЂўР В РЎвЂ“Р В РЎвЂР В РЎвЂќР В Р’Вµ, API Р В РЎвЂР В Р’В»Р В РЎвЂ Р РЋРІР‚С™Р В Р’ВµР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚В¦.
// EN: What it does: groupAllocation stores the best per-group exact allocation candidate found during structured solving.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р В Р вЂ Р В Р’В°Р В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РЎвЂќР В Р’В°Р В РЎвЂќ Р В РЎвЂќР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р РЋР вЂљР В Р’В°Р В РЎвЂќР РЋРІР‚С™ Р В РЎвЂР В Р’В»Р В РЎвЂ Р В РЎвЂўР В РЎвЂ”Р В РЎвЂўР РЋР вЂљР В Р вЂ¦Р В Р’В°Р РЋР РЏ Р РЋРІР‚С™Р В РЎвЂўР РЋРІР‚РЋР В РЎвЂќР В Р’В° Р В РўвЂР В Р’В»Р РЋР РЏ Р В РўвЂР РЋР вЂљР РЋРЎвЂњР В РЎвЂ“Р В РЎвЂР РЋРІР‚В¦ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р’ВµР В РІвЂћвЂ“ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°; Р В РЎвЂР В Р’В·Р В РЎВР В Р’ВµР В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР РЋР РЏ Р В Р’В·Р В РўвЂР В Р’ВµР РЋР С“Р РЋР Р‰ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В РЎвЂў Р РЋРІР‚С™Р РЋР вЂљР В Р’ВµР В Р’В±Р РЋРЎвЂњР РЋР вЂ№Р РЋРІР‚С™ Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В РЎвЂўР В Р’В¶Р В Р вЂ¦Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂ.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type groupAllocation struct {
	amount     int
	score      int
	quantities []int
	ok         bool
}

// RU: Р В РЎС›Р В РЎвЂР В РЎвЂ” Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `App`.
// EN: Data type `App`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂўР В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р РЋР С“Р РЋРІР‚С™Р РЋР вЂљР РЋРЎвЂњР В РЎвЂќР РЋРІР‚С™Р РЋРЎвЂњР РЋР вЂљР РЋРЎвЂњ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ `App`, Р В РЎвЂќР В РЎвЂўР РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В Р’В°Р РЋР РЏ Р РЋРЎвЂњР РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р вЂ Р РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В Р вЂ  Р В Р’В±Р В РЎвЂР В Р’В·Р В Р вЂ¦Р В Р’ВµР РЋР С“-Р В Р’В»Р В РЎвЂўР В РЎвЂ“Р В РЎвЂР В РЎвЂќР В Р’Вµ, API Р В РЎвЂР В Р’В»Р В РЎвЂ Р РЋРІР‚С™Р В Р’ВµР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚В¦.
// EN: What it does: App owns application state, the database handle and the current in-memory session.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р В Р вЂ Р В Р’В°Р В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РЎвЂќР В Р’В°Р В РЎвЂќ Р В РЎвЂќР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р РЋР вЂљР В Р’В°Р В РЎвЂќР РЋРІР‚С™ Р В РЎвЂР В Р’В»Р В РЎвЂ Р В РЎвЂўР В РЎвЂ”Р В РЎвЂўР РЋР вЂљР В Р вЂ¦Р В Р’В°Р РЋР РЏ Р РЋРІР‚С™Р В РЎвЂўР РЋРІР‚РЋР В РЎвЂќР В Р’В° Р В РўвЂР В Р’В»Р РЋР РЏ Р В РўвЂР РЋР вЂљР РЋРЎвЂњР В РЎвЂ“Р В РЎвЂР РЋРІР‚В¦ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В Р’ВµР В РІвЂћвЂ“ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°; Р В РЎвЂР В Р’В·Р В РЎВР В Р’ВµР В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР РЋР РЏ Р В Р’В·Р В РўвЂР В Р’ВµР РЋР С“Р РЋР Р‰ Р РЋРІР‚РЋР В Р’В°Р РЋР С“Р РЋРІР‚С™Р В РЎвЂў Р РЋРІР‚С™Р РЋР вЂљР В Р’ВµР В Р’В±Р РЋРЎвЂњР РЋР вЂ№Р РЋРІР‚С™ Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В РЎвЂўР В Р’В¶Р В Р вЂ¦Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂ.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type App struct {
	ctx            context.Context
	db             *sql.DB
	mu             sync.RWMutex
	currentSession *User
}

// RU: Р В Р’В¤Р РЋРЎвЂњР В Р вЂ¦Р В РЎвЂќР РЋРІР‚В Р В РЎвЂР РЋР РЏ `NewApp`.
// EN: Function `NewApp`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В Р вЂ Р РЋРІР‚в„–Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В Р вЂ¦Р РЋР РЏР В Р’ВµР РЋРІР‚С™ Р В Р вЂ Р РЋР С“Р В РЎвЂ”Р В РЎвЂўР В РЎВР В РЎвЂўР В РЎвЂ“Р В Р’В°Р РЋРІР‚С™Р В Р’ВµР В Р’В»Р РЋР Р‰Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р РЋР вЂљР В Р’ВµР В РЎвЂўР В Р’В±Р РЋР вЂљР В Р’В°Р В Р’В·Р В РЎвЂўР В Р вЂ Р В Р’В°Р В Р вЂ¦Р В РЎвЂР В Р’Вµ, Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР В РЎвЂќР РЋРЎвЂњ Р В РЎвЂР В Р’В»Р В РЎвЂ Р В РЎвЂ”Р В РЎвЂўР В РўвЂР В РЎвЂ“Р В РЎвЂўР РЋРІР‚С™Р В РЎвЂўР В Р вЂ Р В РЎвЂќР РЋРЎвЂњ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦.
// EN: What it does: NewApp prepares the application object, resolves the working database file and initializes schema/data.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р В Р вЂ Р В Р’В°Р В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РўвЂР В Р’В»Р РЋР РЏ Р РЋРЎвЂњР РЋР С“Р РЋРІР‚С™Р В РЎвЂўР В РІвЂћвЂ“Р РЋРІР‚РЋР В РЎвЂР В Р вЂ Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂ Р В Р’В»Р В РЎвЂўР В РЎвЂ“Р В РЎвЂР В РЎвЂќР В РЎвЂ; Р В РЎВР В РЎвЂўР В Р’В¶Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂР РЋР С“Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р РЋР Р‰Р В Р’В·Р В РЎвЂўР В Р вЂ Р В Р’В°Р РЋРІР‚С™Р РЋР Р‰Р РЋР С“Р РЋР РЏ Р РЋР С“Р РЋР вЂљР В Р’В°Р В Р’В·Р РЋРЎвЂњ Р В Р вЂ  Р В Р вЂ¦Р В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В Р’В»Р РЋР Р‰Р В РЎвЂќР В РЎвЂР РЋРІР‚В¦ Р В РЎВР В Р’ВµР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚В¦; Р В РЎвЂР В Р’В·Р В РЎВР В Р’ВµР В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР РЋР РЏ Р РЋР С“Р РЋРІР‚С™Р В РЎвЂўР В РЎвЂР РЋРІР‚С™ Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р РЋРІР‚С™Р РЋР Р‰ Р В РЎвЂўР РЋР С“Р В РЎвЂўР В Р’В·Р В Р вЂ¦Р В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р В РЎвЂў.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
