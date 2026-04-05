package main

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
	RoleAdmin:                     {Department: DepartmentGlobal, Level: 5, Label: "Р С’Р Т‘Р СР С‘Р Р…Р С‘РЎРѓРЎвЂљРЎР‚Р В°РЎвЂљР С•РЎР‚"},
	RoleGlobalDirector:            {Department: DepartmentGlobal, Level: 4, Label: "Р вЂњР ВµР Р…Р ВµРЎР‚Р В°Р В»РЎРЉР Р…РЎвЂ№Р в„– Р Т‘Р С‘РЎР‚Р ВµР С”РЎвЂљР С•РЎР‚"},
	RoleExecutiveDirector:         {Department: DepartmentGlobal, Level: 4, Label: "Р ВРЎРѓР С—Р С•Р В»Р Р…Р С‘РЎвЂљР ВµР В»РЎРЉР Р…РЎвЂ№Р в„– Р Т‘Р С‘РЎР‚Р ВµР С”РЎвЂљР С•РЎР‚"},
	RoleTechnicalDirector:         {Department: DepartmentGlobal, Level: 4, Label: "Р СћР ВµРЎвЂ¦Р Р…Р С‘РЎвЂЎР ВµРЎРѓР С”Р С‘Р в„– Р Т‘Р С‘РЎР‚Р ВµР С”РЎвЂљР С•РЎР‚"},
	RoleSupportHead:               {Department: DepartmentSupport, Level: 3, Label: "Р В РЎС“Р С”Р С•Р Р†Р С•Р Т‘Р С‘РЎвЂљР ВµР В»РЎРЉ Р СћР ВµРЎвЂ¦. Р СџР С•Р Т‘Р Т‘Р ВµРЎР‚Р В¶Р С”Р С‘"},
	RoleSupportSysadmin:           {Department: DepartmentSupport, Level: 3, Label: "Р РЋР С‘РЎРѓРЎвЂљР ВµР СР Р…РЎвЂ№Р в„– Р В°Р Т‘Р СР С‘Р Р…Р С‘РЎРѓРЎвЂљРЎР‚Р В°РЎвЂљР С•РЎР‚"},
	RoleSupportSenior:             {Department: DepartmentSupport, Level: 2, Label: "Р РЋРЎвЂљР В°РЎР‚РЎв‚¬Р С‘Р в„– РЎРѓР С—Р ВµРЎвЂ Р С‘Р В°Р В»Р С‘РЎРѓРЎвЂљ РЎвЂљР ВµРЎвЂ¦Р С—Р С•Р Т‘Р Т‘Р ВµРЎР‚Р В¶Р С”Р С‘"},
	RoleSupportEmployee:           {Department: DepartmentSupport, Level: 1, Label: "Р РЋР С—Р ВµРЎвЂ Р С‘Р В°Р В»Р С‘РЎРѓРЎвЂљ РЎвЂљР ВµРЎвЂ¦Р С—Р С•Р Т‘Р Т‘Р ВµРЎР‚Р В¶Р С”Р С‘"},
	RoleTechnicalHead:             {Department: DepartmentTechnical, Level: 3, Label: "Р В РЎС“Р С”Р С•Р Р†Р С•Р Т‘Р С‘РЎвЂљР ВµР В»РЎРЉ РЎвЂљР ВµРЎвЂ¦Р Р…Р С‘РЎвЂЎР ВµРЎРѓР С”Р С•Р С–Р С• Р С•РЎвЂљР Т‘Р ВµР В»Р В°"},
	RoleTechnicalSenior:           {Department: DepartmentTechnical, Level: 2, Label: "Р РЋРЎвЂљР В°РЎР‚РЎв‚¬Р С‘Р в„– РЎвЂљР ВµРЎвЂ¦Р Р…Р С‘Р С”"},
	RoleTechnicalEmployee:         {Department: DepartmentTechnical, Level: 1, Label: "Р СћР ВµРЎвЂ¦Р Р…Р С‘Р С”"},
	RoleTelecomDirector:           {Department: DepartmentTelecom, Level: 3, Label: "Р вЂќР С‘РЎР‚Р ВµР С”РЎвЂљР С•РЎР‚ Р С—Р С• РЎРѓРЎвЂљРЎР‚Р С•Р С‘РЎвЂљР ВµР В»РЎРЉРЎРѓРЎвЂљР Р†РЎС“"},
	RoleTelecomHead:               {Department: DepartmentTelecom, Level: 3, Label: "Р В РЎС“Р С”Р С•Р Р†Р С•Р Т‘Р С‘РЎвЂљР ВµР В»РЎРЉ РЎРѓРЎвЂљРЎР‚Р С•Р С‘РЎвЂљР ВµР В»РЎРЉР Р…Р С•Р С–Р С• Р С•РЎвЂљР Т‘Р ВµР В»Р В°"},
	RoleTelecomSeniorVOLS:         {Department: DepartmentTelecom, Level: 2, Label: "Р РЋРЎвЂљР В°РЎР‚РЎв‚¬Р С‘Р в„– Р СР С•Р Р…РЎвЂљР В°Р В¶Р Р…Р С‘Р С” Р вЂ™Р С›Р вЂєР РЋ"},
	RoleTelecomSeniorLVS:          {Department: DepartmentTelecom, Level: 2, Label: "Р РЋРЎвЂљР В°РЎР‚РЎв‚¬Р С‘Р в„– Р СР С•Р Р…РЎвЂљР В°Р В¶Р Р…Р С‘Р С” Р вЂєР вЂ™Р РЋ"},
	RoleTelecomEmployeeVOLS:       {Department: DepartmentTelecom, Level: 1, Label: "Р СљР С•Р Р…РЎвЂљР В°Р В¶Р Р…Р С‘Р С” Р вЂ™Р С›Р вЂєР РЋ"},
	RoleTelecomEmployeeLVS:        {Department: DepartmentTelecom, Level: 1, Label: "Р СљР С•Р Р…РЎвЂљР В°Р В¶Р Р…Р С‘Р С” Р вЂєР вЂ™Р РЋ"},
	RoleSKUDHead:                  {Department: DepartmentSKUD, Level: 3, Label: "Р В РЎС“Р С”Р С•Р Р†Р С•Р Т‘Р С‘РЎвЂљР ВµР В»РЎРЉ Р С•РЎвЂљР Т‘Р ВµР В»Р В° РЎвЂљР ВµРЎвЂ¦Р Р…Р С‘РЎвЂЎР ВµРЎРѓР С”Р С•Р С–Р С• Р С•Р В±РЎРѓР В»РЎС“Р В¶Р С‘Р Р†Р В°Р Р…Р С‘РЎРЏ Р РЋР С™Р Р€Р вЂќ"},
	RoleSKUDProjectManager:        {Department: DepartmentSKUD, Level: 2, Label: "Р СљР ВµР Р…Р ВµР Т‘Р В¶Р ВµРЎР‚ Р С—РЎР‚Р С•Р ВµР С”РЎвЂљР С•Р Р† Р РЋР С™Р Р€Р вЂќ"},
	RoleSKUDSeniorService:         {Department: DepartmentSKUD, Level: 2, Label: "Р РЋРЎвЂљР В°РЎР‚РЎв‚¬Р С‘Р в„– РЎРѓР ВµРЎР‚Р Р†Р С‘РЎРѓР Р…РЎвЂ№Р в„– Р С‘Р Р…Р В¶Р ВµР Р…Р ВµРЎР‚ Р РЋР С™Р Р€Р вЂќ"},
	RoleSKUDSeniorInstaller:       {Department: DepartmentSKUD, Level: 2, Label: "Р РЋРЎвЂљР В°РЎР‚РЎв‚¬Р С‘Р в„– Р СР С•Р Р…РЎвЂљР В°Р В¶Р Р…Р С‘Р С” Р РЋР С™Р Р€Р вЂќ"},
	RoleSKUDServiceEngineer:       {Department: DepartmentSKUD, Level: 1, Label: "Р РЋР ВµРЎР‚Р Р†Р С‘РЎРѓР Р…РЎвЂ№Р в„– Р С‘Р Р…Р В¶Р ВµР Р…Р ВµРЎР‚ Р РЋР С™Р Р€Р вЂќ"},
	RoleSKUDInstaller:             {Department: DepartmentSKUD, Level: 1, Label: "Р СљР С•Р Р…РЎвЂљР В°Р В¶Р Р…Р С‘Р С” Р РЋР С™Р Р€Р вЂќ"},
	RoleApprovalHead:              {Department: DepartmentApproval, Level: 3, Label: "Р В РЎС“Р С”Р С•Р Р†Р С•Р Т‘Р С‘РЎвЂљР ВµР В»РЎРЉ РЎРѓР С•Р С–Р В»Р В°РЎРѓР С•Р Р†Р В°Р Р…Р С‘РЎРЏ"},
	RoleApprovalSenior:            {Department: DepartmentApproval, Level: 2, Label: "Р РЋРЎвЂљР В°РЎР‚РЎв‚¬Р С‘Р в„– Р СР ВµР Р…Р ВµР Т‘Р В¶Р ВµРЎР‚ РЎРѓР С•Р С–Р В»Р В°РЎРѓР С•Р Р†Р В°Р Р…Р С‘РЎРЏ"},
	RoleApprovalEmployee:          {Department: DepartmentApproval, Level: 1, Label: "Р СљР ВµР Р…Р ВµР Т‘Р В¶Р ВµРЎР‚ Р С—Р С• РЎРѓР С•Р С–Р В»Р В°РЎРѓР С•Р Р†Р В°Р Р…Р С‘РЎР‹"},
	RoleMarketingHead:             {Department: DepartmentMarketing, Level: 3, Label: "Р В РЎС“Р С”Р С•Р Р†Р С•Р Т‘Р С‘РЎвЂљР ВµР В»РЎРЉ Р С•РЎвЂљР Т‘Р ВµР В»Р В° РЎР‚Р ВµР С”Р В»Р В°Р СРЎвЂ№ Р С‘ Р СР В°РЎР‚Р С”Р ВµРЎвЂљР С‘Р Р…Р С–Р В°"},
	RoleMarketingCourier:          {Department: DepartmentMarketing, Level: 1, Label: "Р С™РЎС“РЎР‚РЎРЉР ВµРЎР‚"},
	RoleCommercialDirector:        {Department: DepartmentCommercial, Level: 4, Label: "Р В РЎС“Р С”Р С•Р Р†Р С•Р Т‘Р С‘РЎвЂљР ВµР В»РЎРЉ Р С•РЎвЂљР Т‘Р ВµР В»Р В° РЎР‚Р ВµР С”Р В»Р В°Р СРЎвЂ№ Р С‘ Р СР В°РЎР‚Р С”Р ВµРЎвЂљР С‘Р Р…Р С–Р В°"},
	RoleCommercialSubscriberHead:  {Department: DepartmentCommercial, Level: 3, Label: "Р В РЎС“Р С”Р С•Р Р†Р С•Р Т‘Р С‘РЎвЂљР ВµР В»РЎРЉ Р В°Р В±Р С•Р Р…Р ВµР Р…РЎвЂљРЎРѓР С”Р С•Р С–Р С• Р С•РЎвЂљР Т‘Р ВµР В»Р В°"},
	RoleCommercialActiveSalesHead: {Department: DepartmentCommercial, Level: 3, Label: "Р СљР ВµР Р…Р ВµР Т‘Р В¶Р ВµРЎР‚ Р В°Р С”РЎвЂљР С‘Р Р†Р Р…РЎвЂ№РЎвЂ¦ Р С—РЎР‚Р С•Р Т‘Р В°Р В¶"},
	RoleCommercialSeniorMRK:       {Department: DepartmentCommercial, Level: 2, Label: "Р РЋРЎвЂљР В°РЎР‚РЎв‚¬Р С‘Р в„– Р СљР В Р С™"},
	RoleCommercialSeniorMRYU:      {Department: DepartmentCommercial, Level: 2, Label: "Р РЋРЎвЂљР В°РЎР‚РЎв‚¬Р С‘Р в„– Р СљР В Р В®"},
	RoleCommercialEmployeeMRK:     {Department: DepartmentCommercial, Level: 1, Label: "Р СљР В Р С™"},
	RoleCommercialEmployeeMRYU:    {Department: DepartmentCommercial, Level: 1, Label: "Р СљР В Р В®"},
	RoleFinanceHead:               {Department: DepartmentFinance, Level: 3, Label: "Р вЂњР В». Р В±РЎС“РЎвЂ¦Р С–Р В°Р В»РЎвЂљР ВµРЎР‚"},
	RoleFinanceEmployee:           {Department: DepartmentFinance, Level: 1, Label: "Р СџР С•Р СР С•РЎвЂ°Р Р…Р С‘Р С” Р В±РЎС“РЎвЂ¦Р С–Р В°Р В»РЎвЂљР ВµРЎР‚Р В°"},
	RoleLegalEmployee:             {Department: DepartmentLegal, Level: 1, Label: "Р В®РЎР‚Р С‘РЎРѓРЎвЂљ"},
	RoleDevelopmentHead:           {Department: DepartmentDevelopment, Level: 3, Label: "Р В РЎС“Р С”Р С•Р Р†Р С•Р Т‘Р С‘РЎвЂљР ВµР В»РЎРЉ Р С–РЎР‚РЎС“Р С—Р С—РЎвЂ№ РЎР‚Р В°Р В·РЎР‚Р В°Р В±Р С•РЎвЂљР С”Р С‘"},
	RoleDevelopmentSenior:         {Department: DepartmentDevelopment, Level: 2, Label: "Р РЋРЎвЂљР В°РЎР‚РЎв‚¬Р С‘Р в„– РЎР‚Р В°Р В·РЎР‚Р В°Р В±Р С•РЎвЂљРЎвЂЎР С‘Р С”"},
	RoleDevelopmentEmployee:       {Department: DepartmentDevelopment, Level: 1, Label: "Р В Р В°Р В·РЎР‚Р В°Р В±Р С•РЎвЂљРЎвЂЎР С‘Р С”"},
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

// RU: Р СћР С‘Р С— Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `Service`.
// EN: Data type `Service`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљРЎС“РЎР‚РЎС“ Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `Service`, Р С”Р С•РЎвЂљР С•РЎР‚Р В°РЎРЏ РЎС“РЎвЂЎР В°РЎРѓРЎвЂљР Р†РЎС“Р ВµРЎвЂљ Р Р† Р В±Р С‘Р В·Р Р…Р ВµРЎРѓ-Р В»Р С•Р С–Р С‘Р С”Р Вµ, API Р С‘Р В»Р С‘ РЎвЂљР ВµРЎРѓРЎвЂљР В°РЎвЂ¦.
// EN: What it does: Service describes one billable position available to the current user: pricing, unit, ownership and allocation settings.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р С”Р В°Р С” Р С”Р С•Р Р…РЎвЂљРЎР‚Р В°Р С”РЎвЂљ Р С‘Р В»Р С‘ Р С•Р С—Р С•РЎР‚Р Р…Р В°РЎРЏ РЎвЂљР С•РЎвЂЎР С”Р В° Р Т‘Р В»РЎРЏ Р Т‘РЎР‚РЎС“Р С–Р С‘РЎвЂ¦ РЎвЂЎР В°РЎРѓРЎвЂљР ВµР в„– Р С—РЎР‚Р С•Р ВµР С”РЎвЂљР В°; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ Р В·Р Т‘Р ВµРЎРѓРЎРЉ РЎвЂЎР В°РЎРѓРЎвЂљР С• РЎвЂљРЎР‚Р ВµР В±РЎС“РЎР‹РЎвЂљ Р С•РЎРѓРЎвЂљР С•РЎР‚Р С•Р В¶Р Р…Р С•РЎРѓРЎвЂљР С‘.
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

// RU: Р СћР С‘Р С— Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `UpsertServiceRequest`.
// EN: Data type `UpsertServiceRequest`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљРЎС“РЎР‚РЎС“ Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `UpsertServiceRequest`, Р С”Р С•РЎвЂљР С•РЎР‚Р В°РЎРЏ РЎС“РЎвЂЎР В°РЎРѓРЎвЂљР Р†РЎС“Р ВµРЎвЂљ Р Р† Р В±Р С‘Р В·Р Р…Р ВµРЎРѓ-Р В»Р С•Р С–Р С‘Р С”Р Вµ, API Р С‘Р В»Р С‘ РЎвЂљР ВµРЎРѓРЎвЂљР В°РЎвЂ¦.
// EN: What it does: UpsertServiceRequest is the payload used by the frontend when a user creates or edits a service definition.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р С”Р В°Р С” Р С”Р С•Р Р…РЎвЂљРЎР‚Р В°Р С”РЎвЂљ Р С‘Р В»Р С‘ Р С•Р С—Р С•РЎР‚Р Р…Р В°РЎРЏ РЎвЂљР С•РЎвЂЎР С”Р В° Р Т‘Р В»РЎРЏ Р Т‘РЎР‚РЎС“Р С–Р С‘РЎвЂ¦ РЎвЂЎР В°РЎРѓРЎвЂљР ВµР в„– Р С—РЎР‚Р С•Р ВµР С”РЎвЂљР В°; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ Р В·Р Т‘Р ВµРЎРѓРЎРЉ РЎвЂЎР В°РЎРѓРЎвЂљР С• РЎвЂљРЎР‚Р ВµР В±РЎС“РЎР‹РЎвЂљ Р С•РЎРѓРЎвЂљР С•РЎР‚Р С•Р В¶Р Р…Р С•РЎРѓРЎвЂљР С‘.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type UpsertServiceRequest struct {
	ID                int64    `json:"id"`
	Name              string   `json:"name"`
	Unit              string   `json:"unit"`
	Rate              int      `json:"rate"`
	Category          string   `json:"category"`
	AllocationPercent *float64 `json:"allocationPercent"`
}

// RU: Р СћР С‘Р С— Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `User`.
// EN: Data type `User`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљРЎС“РЎР‚РЎС“ Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `User`, Р С”Р С•РЎвЂљР С•РЎР‚Р В°РЎРЏ РЎС“РЎвЂЎР В°РЎРѓРЎвЂљР Р†РЎС“Р ВµРЎвЂљ Р Р† Р В±Р С‘Р В·Р Р…Р ВµРЎРѓ-Р В»Р С•Р С–Р С‘Р С”Р Вµ, API Р С‘Р В»Р С‘ РЎвЂљР ВµРЎРѓРЎвЂљР В°РЎвЂ¦.
// EN: What it does: User stores public account information that can be safely returned to the frontend without password data.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р С”Р В°Р С” Р С”Р С•Р Р…РЎвЂљРЎР‚Р В°Р С”РЎвЂљ Р С‘Р В»Р С‘ Р С•Р С—Р С•РЎР‚Р Р…Р В°РЎРЏ РЎвЂљР С•РЎвЂЎР С”Р В° Р Т‘Р В»РЎРЏ Р Т‘РЎР‚РЎС“Р С–Р С‘РЎвЂ¦ РЎвЂЎР В°РЎРѓРЎвЂљР ВµР в„– Р С—РЎР‚Р С•Р ВµР С”РЎвЂљР В°; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ Р В·Р Т‘Р ВµРЎРѓРЎРЉ РЎвЂЎР В°РЎРѓРЎвЂљР С• РЎвЂљРЎР‚Р ВµР В±РЎС“РЎР‹РЎвЂљ Р С•РЎРѓРЎвЂљР С•РЎР‚Р С•Р В¶Р Р…Р С•РЎРѓРЎвЂљР С‘.
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

// RU: Р СћР С‘Р С— Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `UserWithPassword`.
// EN: Data type `UserWithPassword`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљРЎС“РЎР‚РЎС“ Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `UserWithPassword`, Р С”Р С•РЎвЂљР С•РЎР‚Р В°РЎРЏ РЎС“РЎвЂЎР В°РЎРѓРЎвЂљР Р†РЎС“Р ВµРЎвЂљ Р Р† Р В±Р С‘Р В·Р Р…Р ВµРЎРѓ-Р В»Р С•Р С–Р С‘Р С”Р Вµ, API Р С‘Р В»Р С‘ РЎвЂљР ВµРЎРѓРЎвЂљР В°РЎвЂ¦.
// EN: What it does: UserWithPassword extends a user payload with a plain-text password for creation workflows only.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р С”Р В°Р С” Р С”Р С•Р Р…РЎвЂљРЎР‚Р В°Р С”РЎвЂљ Р С‘Р В»Р С‘ Р С•Р С—Р С•РЎР‚Р Р…Р В°РЎРЏ РЎвЂљР С•РЎвЂЎР С”Р В° Р Т‘Р В»РЎРЏ Р Т‘РЎР‚РЎС“Р С–Р С‘РЎвЂ¦ РЎвЂЎР В°РЎРѓРЎвЂљР ВµР в„– Р С—РЎР‚Р С•Р ВµР С”РЎвЂљР В°; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ Р В·Р Т‘Р ВµРЎРѓРЎРЉ РЎвЂЎР В°РЎРѓРЎвЂљР С• РЎвЂљРЎР‚Р ВµР В±РЎС“РЎР‹РЎвЂљ Р С•РЎРѓРЎвЂљР С•РЎР‚Р С•Р В¶Р Р…Р С•РЎРѓРЎвЂљР С‘.
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

// RU: Р СћР С‘Р С— Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `LoginRequest`.
// EN: Data type `LoginRequest`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљРЎС“РЎР‚РЎС“ Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `LoginRequest`, Р С”Р С•РЎвЂљР С•РЎР‚Р В°РЎРЏ РЎС“РЎвЂЎР В°РЎРѓРЎвЂљР Р†РЎС“Р ВµРЎвЂљ Р Р† Р В±Р С‘Р В·Р Р…Р ВµРЎРѓ-Р В»Р С•Р С–Р С‘Р С”Р Вµ, API Р С‘Р В»Р С‘ РЎвЂљР ВµРЎРѓРЎвЂљР В°РЎвЂ¦.
// EN: What it does: LoginRequest carries credentials from the login form to the backend session logic.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р С”Р В°Р С” Р С”Р С•Р Р…РЎвЂљРЎР‚Р В°Р С”РЎвЂљ Р С‘Р В»Р С‘ Р С•Р С—Р С•РЎР‚Р Р…Р В°РЎРЏ РЎвЂљР С•РЎвЂЎР С”Р В° Р Т‘Р В»РЎРЏ Р Т‘РЎР‚РЎС“Р С–Р С‘РЎвЂ¦ РЎвЂЎР В°РЎРѓРЎвЂљР ВµР в„– Р С—РЎР‚Р С•Р ВµР С”РЎвЂљР В°; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ Р В·Р Т‘Р ВµРЎРѓРЎРЉ РЎвЂЎР В°РЎРѓРЎвЂљР С• РЎвЂљРЎР‚Р ВµР В±РЎС“РЎР‹РЎвЂљ Р С•РЎРѓРЎвЂљР С•РЎР‚Р С•Р В¶Р Р…Р С•РЎРѓРЎвЂљР С‘.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RU: Р СћР С‘Р С— Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `SessionState`.
// EN: Data type `SessionState`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљРЎС“РЎР‚РЎС“ Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `SessionState`, Р С”Р С•РЎвЂљР С•РЎР‚Р В°РЎРЏ РЎС“РЎвЂЎР В°РЎРѓРЎвЂљР Р†РЎС“Р ВµРЎвЂљ Р Р† Р В±Р С‘Р В·Р Р…Р ВµРЎРѓ-Р В»Р С•Р С–Р С‘Р С”Р Вµ, API Р С‘Р В»Р С‘ РЎвЂљР ВµРЎРѓРЎвЂљР В°РЎвЂ¦.
// EN: What it does: SessionState is the frontend-facing snapshot of the current authenticated user and their capabilities.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р С”Р В°Р С” Р С”Р С•Р Р…РЎвЂљРЎР‚Р В°Р С”РЎвЂљ Р С‘Р В»Р С‘ Р С•Р С—Р С•РЎР‚Р Р…Р В°РЎРЏ РЎвЂљР С•РЎвЂЎР С”Р В° Р Т‘Р В»РЎРЏ Р Т‘РЎР‚РЎС“Р С–Р С‘РЎвЂ¦ РЎвЂЎР В°РЎРѓРЎвЂљР ВµР в„– Р С—РЎР‚Р С•Р ВµР С”РЎвЂљР В°; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ Р В·Р Т‘Р ВµРЎРѓРЎРЉ РЎвЂЎР В°РЎРѓРЎвЂљР С• РЎвЂљРЎР‚Р ВµР В±РЎС“РЎР‹РЎвЂљ Р С•РЎРѓРЎвЂљР С•РЎР‚Р С•Р В¶Р Р…Р С•РЎРѓРЎвЂљР С‘.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type SessionState struct {
	Authenticated bool   `json:"authenticated"`
	User          *User  `json:"user,omitempty"`
	CanManage     bool   `json:"canManage"`
	CanAdmin      bool   `json:"canAdmin"`
	CanModerate   bool   `json:"canModerate"`
	Message       string `json:"message,omitempty"`
}

// RU: Р СћР С‘Р С— Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `CalculationRequest`.
// EN: Data type `CalculationRequest`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљРЎС“РЎР‚РЎС“ Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `CalculationRequest`, Р С”Р С•РЎвЂљР С•РЎР‚Р В°РЎРЏ РЎС“РЎвЂЎР В°РЎРѓРЎвЂљР Р†РЎС“Р ВµРЎвЂљ Р Р† Р В±Р С‘Р В·Р Р…Р ВµРЎРѓ-Р В»Р С•Р С–Р С‘Р С”Р Вµ, API Р С‘Р В»Р С‘ РЎвЂљР ВµРЎРѓРЎвЂљР В°РЎвЂ¦.
// EN: What it does: CalculationRequest contains the target amount and per-service weights used to build a calculation.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р С”Р В°Р С” Р С”Р С•Р Р…РЎвЂљРЎР‚Р В°Р С”РЎвЂљ Р С‘Р В»Р С‘ Р С•Р С—Р С•РЎР‚Р Р…Р В°РЎРЏ РЎвЂљР С•РЎвЂЎР С”Р В° Р Т‘Р В»РЎРЏ Р Т‘РЎР‚РЎС“Р С–Р С‘РЎвЂ¦ РЎвЂЎР В°РЎРѓРЎвЂљР ВµР в„– Р С—РЎР‚Р С•Р ВµР С”РЎвЂљР В°; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ Р В·Р Т‘Р ВµРЎРѓРЎРЉ РЎвЂЎР В°РЎРѓРЎвЂљР С• РЎвЂљРЎР‚Р ВµР В±РЎС“РЎР‹РЎвЂљ Р С•РЎРѓРЎвЂљР С•РЎР‚Р С•Р В¶Р Р…Р С•РЎРѓРЎвЂљР С‘.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type CalculationRequest struct {
	TargetAmount int            `json:"targetAmount"`
	Weights      map[string]int `json:"weights"`
}

// RU: Р СћР С‘Р С— Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `CalculationItem`.
// EN: Data type `CalculationItem`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљРЎС“РЎР‚РЎС“ Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `CalculationItem`, Р С”Р С•РЎвЂљР С•РЎР‚Р В°РЎРЏ РЎС“РЎвЂЎР В°РЎРѓРЎвЂљР Р†РЎС“Р ВµРЎвЂљ Р Р† Р В±Р С‘Р В·Р Р…Р ВµРЎРѓ-Р В»Р С•Р С–Р С‘Р С”Р Вµ, API Р С‘Р В»Р С‘ РЎвЂљР ВµРЎРѓРЎвЂљР В°РЎвЂ¦.
// EN: What it does: CalculationItem is one line of the generated or archived calculation with both business and UI-oriented fields.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р С”Р В°Р С” Р С”Р С•Р Р…РЎвЂљРЎР‚Р В°Р С”РЎвЂљ Р С‘Р В»Р С‘ Р С•Р С—Р С•РЎР‚Р Р…Р В°РЎРЏ РЎвЂљР С•РЎвЂЎР С”Р В° Р Т‘Р В»РЎРЏ Р Т‘РЎР‚РЎС“Р С–Р С‘РЎвЂ¦ РЎвЂЎР В°РЎРѓРЎвЂљР ВµР в„– Р С—РЎР‚Р С•Р ВµР С”РЎвЂљР В°; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ Р В·Р Т‘Р ВµРЎРѓРЎРЉ РЎвЂЎР В°РЎРѓРЎвЂљР С• РЎвЂљРЎР‚Р ВµР В±РЎС“РЎР‹РЎвЂљ Р С•РЎРѓРЎвЂљР С•РЎР‚Р С•Р В¶Р Р…Р С•РЎРѓРЎвЂљР С‘.
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

// RU: Р СћР С‘Р С— Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `CalculationResult`.
// EN: Data type `CalculationResult`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљРЎС“РЎР‚РЎС“ Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `CalculationResult`, Р С”Р С•РЎвЂљР С•РЎР‚Р В°РЎРЏ РЎС“РЎвЂЎР В°РЎРѓРЎвЂљР Р†РЎС“Р ВµРЎвЂљ Р Р† Р В±Р С‘Р В·Р Р…Р ВµРЎРѓ-Р В»Р С•Р С–Р С‘Р С”Р Вµ, API Р С‘Р В»Р С‘ РЎвЂљР ВµРЎРѓРЎвЂљР В°РЎвЂ¦.
// EN: What it does: CalculationResult is the full response returned after a calculation attempt, including totals and metadata.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р С”Р В°Р С” Р С”Р С•Р Р…РЎвЂљРЎР‚Р В°Р С”РЎвЂљ Р С‘Р В»Р С‘ Р С•Р С—Р С•РЎР‚Р Р…Р В°РЎРЏ РЎвЂљР С•РЎвЂЎР С”Р В° Р Т‘Р В»РЎРЏ Р Т‘РЎР‚РЎС“Р С–Р С‘РЎвЂ¦ РЎвЂЎР В°РЎРѓРЎвЂљР ВµР в„– Р С—РЎР‚Р С•Р ВµР С”РЎвЂљР В°; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ Р В·Р Т‘Р ВµРЎРѓРЎРЉ РЎвЂЎР В°РЎРѓРЎвЂљР С• РЎвЂљРЎР‚Р ВµР В±РЎС“РЎР‹РЎвЂљ Р С•РЎРѓРЎвЂљР С•РЎР‚Р С•Р В¶Р Р…Р С•РЎРѓРЎвЂљР С‘.
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

// RU: Р СћР С‘Р С— Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `SaveCalculationRequest`.
// EN: Data type `SaveCalculationRequest`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљРЎС“РЎР‚РЎС“ Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `SaveCalculationRequest`, Р С”Р С•РЎвЂљР С•РЎР‚Р В°РЎРЏ РЎС“РЎвЂЎР В°РЎРѓРЎвЂљР Р†РЎС“Р ВµРЎвЂљ Р Р† Р В±Р С‘Р В·Р Р…Р ВµРЎРѓ-Р В»Р С•Р С–Р С‘Р С”Р Вµ, API Р С‘Р В»Р С‘ РЎвЂљР ВµРЎРѓРЎвЂљР В°РЎвЂ¦.
// EN: What it does: SaveCalculationRequest is sent when the frontend persists the current calculation into the archive.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р С”Р В°Р С” Р С”Р С•Р Р…РЎвЂљРЎР‚Р В°Р С”РЎвЂљ Р С‘Р В»Р С‘ Р С•Р С—Р С•РЎР‚Р Р…Р В°РЎРЏ РЎвЂљР С•РЎвЂЎР С”Р В° Р Т‘Р В»РЎРЏ Р Т‘РЎР‚РЎС“Р С–Р С‘РЎвЂ¦ РЎвЂЎР В°РЎРѓРЎвЂљР ВµР в„– Р С—РЎР‚Р С•Р ВµР С”РЎвЂљР В°; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ Р В·Р Т‘Р ВµРЎРѓРЎРЉ РЎвЂЎР В°РЎРѓРЎвЂљР С• РЎвЂљРЎР‚Р ВµР В±РЎС“РЎР‹РЎвЂљ Р С•РЎРѓРЎвЂљР С•РЎР‚Р С•Р В¶Р Р…Р С•РЎРѓРЎвЂљР С‘.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type SaveCalculationRequest struct {
	Title        string            `json:"title"`
	TargetAmount int               `json:"targetAmount"`
	Items        []CalculationItem `json:"items"`
}

// RU: Р СћР С‘Р С— Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `SavedCalculation`.
// EN: Data type `SavedCalculation`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљРЎС“РЎР‚РЎС“ Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `SavedCalculation`, Р С”Р С•РЎвЂљР С•РЎР‚Р В°РЎРЏ РЎС“РЎвЂЎР В°РЎРѓРЎвЂљР Р†РЎС“Р ВµРЎвЂљ Р Р† Р В±Р С‘Р В·Р Р…Р ВµРЎРѓ-Р В»Р С•Р С–Р С‘Р С”Р Вµ, API Р С‘Р В»Р С‘ РЎвЂљР ВµРЎРѓРЎвЂљР В°РЎвЂ¦.
// EN: What it does: SavedCalculation represents one archived calculation entry as stored in SQLite and shown in history.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р С”Р В°Р С” Р С”Р С•Р Р…РЎвЂљРЎР‚Р В°Р С”РЎвЂљ Р С‘Р В»Р С‘ Р С•Р С—Р С•РЎР‚Р Р…Р В°РЎРЏ РЎвЂљР С•РЎвЂЎР С”Р В° Р Т‘Р В»РЎРЏ Р Т‘РЎР‚РЎС“Р С–Р С‘РЎвЂ¦ РЎвЂЎР В°РЎРѓРЎвЂљР ВµР в„– Р С—РЎР‚Р С•Р ВµР С”РЎвЂљР В°; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ Р В·Р Т‘Р ВµРЎРѓРЎРЉ РЎвЂЎР В°РЎРѓРЎвЂљР С• РЎвЂљРЎР‚Р ВµР В±РЎС“РЎР‹РЎвЂљ Р С•РЎРѓРЎвЂљР С•РЎР‚Р С•Р В¶Р Р…Р С•РЎРѓРЎвЂљР С‘.
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

// RU: Р СћР С‘Р С— Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `CopyArchiveServicesResult`.
// EN: Data type `CopyArchiveServicesResult`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ РЎР‚Р ВµР В·РЎС“Р В»РЎРЉРЎвЂљР В°РЎвЂљ Р С”Р С•Р С—Р С‘РЎР‚Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ РЎС“РЎРѓР В»РЎС“Р С– Р С‘Р В· Р В°РЎР‚РЎвЂ¦Р С‘Р Р†Р Р…Р С•Р С–Р С• РЎР‚Р В°РЎРѓРЎвЂЎРЎвЂРЎвЂљР В° Р Р† РЎРѓР С—Р С‘РЎРѓР С•Р С” РЎС“РЎРѓР В»РЎС“Р С– Р В°Р Т‘Р СР С‘Р Р…Р С‘РЎРѓРЎвЂљРЎР‚Р В°РЎвЂљР С•РЎР‚Р В°.
// EN: What it does: CopyArchiveServicesResult reports how many services were created or updated after importing from an archived calculation.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р С‘РЎРѓР С—Р С•Р В»РЎРЉР В·РЎС“Р ВµРЎвЂљРЎРѓРЎРЏ РЎвЂљР С•Р В»РЎРЉР С”Р С• Р Р† Р В°Р Т‘Р СР С‘Р Р…РЎРѓР С”Р С•Р С РЎРѓРЎвЂ Р ВµР Р…Р В°РЎР‚Р С‘Р С‘; Р С—Р С•Р СР С•Р С–Р В°Р ВµРЎвЂљ РЎвЂћРЎР‚Р С•Р Р…РЎвЂљР ВµР Р…Р Т‘РЎС“ Р С—Р С•Р С”Р В°Р В·Р В°РЎвЂљРЎРЉ Р С—Р С•Р Р…РЎРЏРЎвЂљР Р…Р С•Р Вµ РЎРѓР С•Р С•Р В±РЎвЂ°Р ВµР Р…Р С‘Р Вµ Р С—Р С•РЎРѓР В»Р Вµ Р С‘Р СР С—Р С•РЎР‚РЎвЂљР В°; Р С•РЎвЂљР Т‘Р ВµР В»РЎРЏР ВµРЎвЂљ РЎРѓРЎвЂљР В°РЎвЂљР С‘РЎРѓРЎвЂљР С‘Р С”РЎС“ Р С•Р С—Р ВµРЎР‚Р В°РЎвЂ Р С‘Р С‘ Р С•РЎвЂљ Р С—Р С•Р В»Р Р…Р С•Р С–Р С• РЎРѓР С—Р С‘РЎРѓР С”Р В° РЎС“РЎРѓР В»РЎС“Р С–.
// EN: Key points: used only in the admin-only archive import flow; lets the frontend show a concise status message; keeps operation stats separate from the full services list.
type CopyArchiveServicesResult struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
}

// RU: Р СћР С‘Р С— Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `AppBootstrap`.
// EN: Data type `AppBootstrap`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљРЎС“РЎР‚РЎС“ Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `AppBootstrap`, Р С”Р С•РЎвЂљР С•РЎР‚Р В°РЎРЏ РЎС“РЎвЂЎР В°РЎРѓРЎвЂљР Р†РЎС“Р ВµРЎвЂљ Р Р† Р В±Р С‘Р В·Р Р…Р ВµРЎРѓ-Р В»Р С•Р С–Р С‘Р С”Р Вµ, API Р С‘Р В»Р С‘ РЎвЂљР ВµРЎРѓРЎвЂљР В°РЎвЂ¦.
// EN: What it does: AppBootstrap aggregates all initial data the frontend needs after startup or refresh.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р С”Р В°Р С” Р С”Р С•Р Р…РЎвЂљРЎР‚Р В°Р С”РЎвЂљ Р С‘Р В»Р С‘ Р С•Р С—Р С•РЎР‚Р Р…Р В°РЎРЏ РЎвЂљР С•РЎвЂЎР С”Р В° Р Т‘Р В»РЎРЏ Р Т‘РЎР‚РЎС“Р С–Р С‘РЎвЂ¦ РЎвЂЎР В°РЎРѓРЎвЂљР ВµР в„– Р С—РЎР‚Р С•Р ВµР С”РЎвЂљР В°; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ Р В·Р Т‘Р ВµРЎРѓРЎРЉ РЎвЂЎР В°РЎРѓРЎвЂљР С• РЎвЂљРЎР‚Р ВµР В±РЎС“РЎР‹РЎвЂљ Р С•РЎРѓРЎвЂљР С•РЎР‚Р С•Р В¶Р Р…Р С•РЎРѓРЎвЂљР С‘.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type AppBootstrap struct {
	Session             SessionState       `json:"session"`
	Services            []Service          `json:"services"`
	Users               []User             `json:"users"`
	SavedCalculations   []SavedCalculation `json:"savedCalculations"`
	DefaultGroupPercent map[string]float64 `json:"defaultGroupPercent"`
}

// RU: Р СћР С‘Р С— Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `allocationState`.
// EN: Data type `allocationState`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљРЎС“РЎР‚РЎС“ Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `allocationState`, Р С”Р С•РЎвЂљР С•РЎР‚Р В°РЎРЏ РЎС“РЎвЂЎР В°РЎРѓРЎвЂљР Р†РЎС“Р ВµРЎвЂљ Р Р† Р В±Р С‘Р В·Р Р…Р ВµРЎРѓ-Р В»Р С•Р С–Р С‘Р С”Р Вµ, API Р С‘Р В»Р С‘ РЎвЂљР ВµРЎРѓРЎвЂљР В°РЎвЂ¦.
// EN: What it does: allocationState is an internal DP cell used while searching for a good quantity distribution.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р С”Р В°Р С” Р С”Р С•Р Р…РЎвЂљРЎР‚Р В°Р С”РЎвЂљ Р С‘Р В»Р С‘ Р С•Р С—Р С•РЎР‚Р Р…Р В°РЎРЏ РЎвЂљР С•РЎвЂЎР С”Р В° Р Т‘Р В»РЎРЏ Р Т‘РЎР‚РЎС“Р С–Р С‘РЎвЂ¦ РЎвЂЎР В°РЎРѓРЎвЂљР ВµР в„– Р С—РЎР‚Р С•Р ВµР С”РЎвЂљР В°; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ Р В·Р Т‘Р ВµРЎРѓРЎРЉ РЎвЂЎР В°РЎРѓРЎвЂљР С• РЎвЂљРЎР‚Р ВµР В±РЎС“РЎР‹РЎвЂљ Р С•РЎРѓРЎвЂљР С•РЎР‚Р С•Р В¶Р Р…Р С•РЎРѓРЎвЂљР С‘.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type allocationState struct {
	distance int
	score    int
	count    int
	prev     int
	idx      int
	ok       bool
}

// RU: Р СћР С‘Р С— Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `groupAllocation`.
// EN: Data type `groupAllocation`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљРЎС“РЎР‚РЎС“ Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `groupAllocation`, Р С”Р С•РЎвЂљР С•РЎР‚Р В°РЎРЏ РЎС“РЎвЂЎР В°РЎРѓРЎвЂљР Р†РЎС“Р ВµРЎвЂљ Р Р† Р В±Р С‘Р В·Р Р…Р ВµРЎРѓ-Р В»Р С•Р С–Р С‘Р С”Р Вµ, API Р С‘Р В»Р С‘ РЎвЂљР ВµРЎРѓРЎвЂљР В°РЎвЂ¦.
// EN: What it does: groupAllocation stores the best per-group exact allocation candidate found during structured solving.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р С”Р В°Р С” Р С”Р С•Р Р…РЎвЂљРЎР‚Р В°Р С”РЎвЂљ Р С‘Р В»Р С‘ Р С•Р С—Р С•РЎР‚Р Р…Р В°РЎРЏ РЎвЂљР С•РЎвЂЎР С”Р В° Р Т‘Р В»РЎРЏ Р Т‘РЎР‚РЎС“Р С–Р С‘РЎвЂ¦ РЎвЂЎР В°РЎРѓРЎвЂљР ВµР в„– Р С—РЎР‚Р С•Р ВµР С”РЎвЂљР В°; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ Р В·Р Т‘Р ВµРЎРѓРЎРЉ РЎвЂЎР В°РЎРѓРЎвЂљР С• РЎвЂљРЎР‚Р ВµР В±РЎС“РЎР‹РЎвЂљ Р С•РЎРѓРЎвЂљР С•РЎР‚Р С•Р В¶Р Р…Р С•РЎРѓРЎвЂљР С‘.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type groupAllocation struct {
	amount     int
	score      int
	quantities []int
	ok         bool
}

// RU: Р СћР С‘Р С— Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `App`.
// EN: Data type `App`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљРЎС“РЎР‚РЎС“ Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ `App`, Р С”Р С•РЎвЂљР С•РЎР‚Р В°РЎРЏ РЎС“РЎвЂЎР В°РЎРѓРЎвЂљР Р†РЎС“Р ВµРЎвЂљ Р Р† Р В±Р С‘Р В·Р Р…Р ВµРЎРѓ-Р В»Р С•Р С–Р С‘Р С”Р Вµ, API Р С‘Р В»Р С‘ РЎвЂљР ВµРЎРѓРЎвЂљР В°РЎвЂ¦.
// EN: What it does: App owns application state, the database handle and the current in-memory session.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р С”Р В°Р С” Р С”Р С•Р Р…РЎвЂљРЎР‚Р В°Р С”РЎвЂљ Р С‘Р В»Р С‘ Р С•Р С—Р С•РЎР‚Р Р…Р В°РЎРЏ РЎвЂљР С•РЎвЂЎР С”Р В° Р Т‘Р В»РЎРЏ Р Т‘РЎР‚РЎС“Р С–Р С‘РЎвЂ¦ РЎвЂЎР В°РЎРѓРЎвЂљР ВµР в„– Р С—РЎР‚Р С•Р ВµР С”РЎвЂљР В°; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ Р В·Р Т‘Р ВµРЎРѓРЎРЉ РЎвЂЎР В°РЎРѓРЎвЂљР С• РЎвЂљРЎР‚Р ВµР В±РЎС“РЎР‹РЎвЂљ Р С•РЎРѓРЎвЂљР С•РЎР‚Р С•Р В¶Р Р…Р С•РЎРѓРЎвЂљР С‘.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type App struct {
	ctx            context.Context
	db             *sql.DB
	mu             sync.RWMutex
	currentSession *User
}

// RU: Р В¤РЎС“Р Р…Р С”РЎвЂ Р С‘РЎРЏ `NewApp`.
// EN: Function `NewApp`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р Р†РЎвЂ№Р С—Р С•Р В»Р Р…РЎРЏР ВµРЎвЂљ Р Р†РЎРѓР С—Р С•Р СР С•Р С–Р В°РЎвЂљР ВµР В»РЎРЉР Р…Р С•Р Вµ Р С—РЎР‚Р ВµР С•Р В±РЎР‚Р В°Р В·Р С•Р Р†Р В°Р Р…Р С‘Р Вµ, Р С—РЎР‚Р С•Р Р†Р ВµРЎР‚Р С”РЎС“ Р С‘Р В»Р С‘ Р С—Р С•Р Т‘Р С–Р С•РЎвЂљР С•Р Р†Р С”РЎС“ Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦.
// EN: What it does: NewApp prepares the application object, resolves the working database file and initializes schema/data.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р Т‘Р В»РЎРЏ РЎС“РЎРѓРЎвЂљР С•Р в„–РЎвЂЎР С‘Р Р†Р С•РЎРѓРЎвЂљР С‘ Р В»Р С•Р С–Р С‘Р С”Р С‘; Р СР С•Р В¶Р ВµРЎвЂљ Р С‘РЎРѓР С—Р С•Р В»РЎРЉР В·Р С•Р Р†Р В°РЎвЂљРЎРЉРЎРѓРЎРЏ РЎРѓРЎР‚Р В°Р В·РЎС“ Р Р† Р Р…Р ВµРЎРѓР С”Р С•Р В»РЎРЉР С”Р С‘РЎвЂ¦ Р СР ВµРЎРѓРЎвЂљР В°РЎвЂ¦; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ РЎРѓРЎвЂљР С•Р С‘РЎвЂљ Р Т‘Р ВµР В»Р В°РЎвЂљРЎРЉ Р С•РЎРѓР С•Р В·Р Р…Р В°Р Р…Р Р…Р С•.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
