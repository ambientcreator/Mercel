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
	RoleAdmin:                     {Department: DepartmentGlobal, Level: 5, Label: "РђРґРјРёРЅРёСЃС‚СЂР°С‚РѕСЂ"},
	RoleGlobalDirector:            {Department: DepartmentGlobal, Level: 4, Label: "Р“РµРЅРµСЂР°Р»СЊРЅС‹Р№ РґРёСЂРµРєС‚РѕСЂ"},
	RoleExecutiveDirector:         {Department: DepartmentGlobal, Level: 4, Label: "РСЃРїРѕР»РЅРёС‚РµР»СЊРЅС‹Р№ РґРёСЂРµРєС‚РѕСЂ"},
	RoleTechnicalDirector:         {Department: DepartmentGlobal, Level: 4, Label: "РўРµС…РЅРёС‡РµСЃРєРёР№ РґРёСЂРµРєС‚РѕСЂ"},
	RoleSupportHead:               {Department: DepartmentSupport, Level: 3, Label: "Р СѓРєРѕРІРѕРґРёС‚РµР»СЊ РўРµС…. РџРѕРґРґРµСЂР¶РєРё"},
	RoleSupportSysadmin:           {Department: DepartmentSupport, Level: 3, Label: "РЎРёСЃС‚РµРјРЅС‹Р№ Р°РґРјРёРЅРёСЃС‚СЂР°С‚РѕСЂ"},
	RoleSupportSenior:             {Department: DepartmentSupport, Level: 2, Label: "РЎС‚Р°СЂС€РёР№ СЃРїРµС†РёР°Р»РёСЃС‚ С‚РµС…РїРѕРґРґРµСЂР¶РєРё"},
	RoleSupportEmployee:           {Department: DepartmentSupport, Level: 1, Label: "РЎРїРµС†РёР°Р»РёСЃС‚ С‚РµС…РїРѕРґРґРµСЂР¶РєРё"},
	RoleTechnicalHead:             {Department: DepartmentTechnical, Level: 3, Label: "Р СѓРєРѕРІРѕРґРёС‚РµР»СЊ С‚РµС…РЅРёС‡РµСЃРєРѕРіРѕ РѕС‚РґРµР»Р°"},
	RoleTechnicalSenior:           {Department: DepartmentTechnical, Level: 2, Label: "РЎС‚Р°СЂС€РёР№ С‚РµС…РЅРёРє"},
	RoleTechnicalEmployee:         {Department: DepartmentTechnical, Level: 1, Label: "РўРµС…РЅРёРє"},
	RoleTelecomDirector:           {Department: DepartmentTelecom, Level: 3, Label: "Р”РёСЂРµРєС‚РѕСЂ РїРѕ СЃС‚СЂРѕРёС‚РµР»СЊСЃС‚РІСѓ"},
	RoleTelecomHead:               {Department: DepartmentTelecom, Level: 3, Label: "Р СѓРєРѕРІРѕРґРёС‚РµР»СЊ СЃС‚СЂРѕРёС‚РµР»СЊРЅРѕРіРѕ РѕС‚РґРµР»Р°"},
	RoleTelecomSeniorVOLS:         {Department: DepartmentTelecom, Level: 2, Label: "РЎС‚Р°СЂС€РёР№ РјРѕРЅС‚Р°Р¶РЅРёРє Р’РћР›РЎ"},
	RoleTelecomSeniorLVS:          {Department: DepartmentTelecom, Level: 2, Label: "РЎС‚Р°СЂС€РёР№ РјРѕРЅС‚Р°Р¶РЅРёРє Р›Р’РЎ"},
	RoleTelecomEmployeeVOLS:       {Department: DepartmentTelecom, Level: 1, Label: "РњРѕРЅС‚Р°Р¶РЅРёРє Р’РћР›РЎ"},
	RoleTelecomEmployeeLVS:        {Department: DepartmentTelecom, Level: 1, Label: "РњРѕРЅС‚Р°Р¶РЅРёРє Р›Р’РЎ"},
	RoleSKUDHead:                  {Department: DepartmentSKUD, Level: 3, Label: "Р СѓРєРѕРІРѕРґРёС‚РµР»СЊ РѕС‚РґРµР»Р° С‚РµС…РЅРёС‡РµСЃРєРѕРіРѕ РѕР±СЃР»СѓР¶РёРІР°РЅРёСЏ РЎРљРЈР”"},
	RoleSKUDProjectManager:        {Department: DepartmentSKUD, Level: 2, Label: "РњРµРЅРµРґР¶РµСЂ РїСЂРѕРµРєС‚РѕРІ РЎРљРЈР”"},
	RoleSKUDSeniorService:         {Department: DepartmentSKUD, Level: 2, Label: "РЎС‚Р°СЂС€РёР№ СЃРµСЂРІРёСЃРЅС‹Р№ РёРЅР¶РµРЅРµСЂ РЎРљРЈР”"},
	RoleSKUDSeniorInstaller:       {Department: DepartmentSKUD, Level: 2, Label: "РЎС‚Р°СЂС€РёР№ РјРѕРЅС‚Р°Р¶РЅРёРє РЎРљРЈР”"},
	RoleSKUDServiceEngineer:       {Department: DepartmentSKUD, Level: 1, Label: "РЎРµСЂРІРёСЃРЅС‹Р№ РёРЅР¶РµРЅРµСЂ РЎРљРЈР”"},
	RoleSKUDInstaller:             {Department: DepartmentSKUD, Level: 1, Label: "РњРѕРЅС‚Р°Р¶РЅРёРє РЎРљРЈР”"},
	RoleApprovalHead:              {Department: DepartmentApproval, Level: 3, Label: "Р СѓРєРѕРІРѕРґРёС‚РµР»СЊ СЃРѕРіР»Р°СЃРѕРІР°РЅРёСЏ"},
	RoleApprovalSenior:            {Department: DepartmentApproval, Level: 2, Label: "РЎС‚Р°СЂС€РёР№ РјРµРЅРµРґР¶РµСЂ СЃРѕРіР»Р°СЃРѕРІР°РЅРёСЏ"},
	RoleApprovalEmployee:          {Department: DepartmentApproval, Level: 1, Label: "РњРµРЅРµРґР¶РµСЂ РїРѕ СЃРѕРіР»Р°СЃРѕРІР°РЅРёСЋ"},
	RoleMarketingHead:             {Department: DepartmentMarketing, Level: 3, Label: "Р СѓРєРѕРІРѕРґРёС‚РµР»СЊ РѕС‚РґРµР»Р° СЂРµРєР»Р°РјС‹ Рё РјР°СЂРєРµС‚РёРЅРіР°"},
	RoleMarketingCourier:          {Department: DepartmentMarketing, Level: 1, Label: "РљСѓСЂСЊРµСЂ"},
	RoleCommercialDirector:        {Department: DepartmentCommercial, Level: 4, Label: "Р СѓРєРѕРІРѕРґРёС‚РµР»СЊ РѕС‚РґРµР»Р° СЂРµРєР»Р°РјС‹ Рё РјР°СЂРєРµС‚РёРЅРіР°"},
	RoleCommercialSubscriberHead:  {Department: DepartmentCommercial, Level: 3, Label: "Р СѓРєРѕРІРѕРґРёС‚РµР»СЊ Р°Р±РѕРЅРµРЅС‚СЃРєРѕРіРѕ РѕС‚РґРµР»Р°"},
	RoleCommercialActiveSalesHead: {Department: DepartmentCommercial, Level: 3, Label: "РњРµРЅРµРґР¶РµСЂ Р°РєС‚РёРІРЅС‹С… РїСЂРѕРґР°Р¶"},
	RoleCommercialSeniorMRK:       {Department: DepartmentCommercial, Level: 2, Label: "РЎС‚Р°СЂС€РёР№ РњР Рљ"},
	RoleCommercialSeniorMRYU:      {Department: DepartmentCommercial, Level: 2, Label: "РЎС‚Р°СЂС€РёР№ РњР Р®"},
	RoleCommercialEmployeeMRK:     {Department: DepartmentCommercial, Level: 1, Label: "РњР Рљ"},
	RoleCommercialEmployeeMRYU:    {Department: DepartmentCommercial, Level: 1, Label: "РњР Р®"},
	RoleFinanceHead:               {Department: DepartmentFinance, Level: 3, Label: "Р“Р». Р±СѓС…РіР°Р»С‚РµСЂ"},
	RoleFinanceEmployee:           {Department: DepartmentFinance, Level: 1, Label: "РџРѕРјРѕС‰РЅРёРє Р±СѓС…РіР°Р»С‚РµСЂР°"},
	RoleLegalEmployee:             {Department: DepartmentLegal, Level: 1, Label: "Р®СЂРёСЃС‚"},
	RoleDevelopmentHead:           {Department: DepartmentDevelopment, Level: 3, Label: "Р СѓРєРѕРІРѕРґРёС‚РµР»СЊ РіСЂСѓРїРїС‹ СЂР°Р·СЂР°Р±РѕС‚РєРё"},
	RoleDevelopmentSenior:         {Department: DepartmentDevelopment, Level: 2, Label: "РЎС‚Р°СЂС€РёР№ СЂР°Р·СЂР°Р±РѕС‚С‡РёРє"},
	RoleDevelopmentEmployee:       {Department: DepartmentDevelopment, Level: 1, Label: "Р Р°Р·СЂР°Р±РѕС‚С‡РёРє"},
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

// RU: РўРёРї РґР°РЅРЅС‹С… `Service`.
// EN: Data type `Service`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РѕРїРёСЃС‹РІР°РµС‚ СЃС‚СЂСѓРєС‚СѓСЂСѓ РґР°РЅРЅС‹С… `Service`, РєРѕС‚РѕСЂР°СЏ СѓС‡Р°СЃС‚РІСѓРµС‚ РІ Р±РёР·РЅРµСЃ-Р»РѕРіРёРєРµ, API РёР»Рё С‚РµСЃС‚Р°С….
// EN: What it does: Service describes one billable position available to the current user: pricing, unit, ownership and allocation settings.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РєР°Рє РєРѕРЅС‚СЂР°РєС‚ РёР»Рё РѕРїРѕСЂРЅР°СЏ С‚РѕС‡РєР° РґР»СЏ РґСЂСѓРіРёС… С‡Р°СЃС‚РµР№ РїСЂРѕРµРєС‚Р°; РёР·РјРµРЅРµРЅРёСЏ Р·РґРµСЃСЊ С‡Р°СЃС‚Рѕ С‚СЂРµР±СѓСЋС‚ РѕСЃС‚РѕСЂРѕР¶РЅРѕСЃС‚Рё.
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

// RU: РўРёРї РґР°РЅРЅС‹С… `UpsertServiceRequest`.
// EN: Data type `UpsertServiceRequest`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РѕРїРёСЃС‹РІР°РµС‚ СЃС‚СЂСѓРєС‚СѓСЂСѓ РґР°РЅРЅС‹С… `UpsertServiceRequest`, РєРѕС‚РѕСЂР°СЏ СѓС‡Р°СЃС‚РІСѓРµС‚ РІ Р±РёР·РЅРµСЃ-Р»РѕРіРёРєРµ, API РёР»Рё С‚РµСЃС‚Р°С….
// EN: What it does: UpsertServiceRequest is the payload used by the frontend when a user creates or edits a service definition.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РєР°Рє РєРѕРЅС‚СЂР°РєС‚ РёР»Рё РѕРїРѕСЂРЅР°СЏ С‚РѕС‡РєР° РґР»СЏ РґСЂСѓРіРёС… С‡Р°СЃС‚РµР№ РїСЂРѕРµРєС‚Р°; РёР·РјРµРЅРµРЅРёСЏ Р·РґРµСЃСЊ С‡Р°СЃС‚Рѕ С‚СЂРµР±СѓСЋС‚ РѕСЃС‚РѕСЂРѕР¶РЅРѕСЃС‚Рё.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type UpsertServiceRequest struct {
	ID                int64    `json:"id"`
	Name              string   `json:"name"`
	Unit              string   `json:"unit"`
	Rate              int      `json:"rate"`
	Category          string   `json:"category"`
	AllocationPercent *float64 `json:"allocationPercent"`
}

// RU: РўРёРї РґР°РЅРЅС‹С… `User`.
// EN: Data type `User`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РѕРїРёСЃС‹РІР°РµС‚ СЃС‚СЂСѓРєС‚СѓСЂСѓ РґР°РЅРЅС‹С… `User`, РєРѕС‚РѕСЂР°СЏ СѓС‡Р°СЃС‚РІСѓРµС‚ РІ Р±РёР·РЅРµСЃ-Р»РѕРіРёРєРµ, API РёР»Рё С‚РµСЃС‚Р°С….
// EN: What it does: User stores public account information that can be safely returned to the frontend without password data.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РєР°Рє РєРѕРЅС‚СЂР°РєС‚ РёР»Рё РѕРїРѕСЂРЅР°СЏ С‚РѕС‡РєР° РґР»СЏ РґСЂСѓРіРёС… С‡Р°СЃС‚РµР№ РїСЂРѕРµРєС‚Р°; РёР·РјРµРЅРµРЅРёСЏ Р·РґРµСЃСЊ С‡Р°СЃС‚Рѕ С‚СЂРµР±СѓСЋС‚ РѕСЃС‚РѕСЂРѕР¶РЅРѕСЃС‚Рё.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt"`
}

// RU: РўРёРї РґР°РЅРЅС‹С… `UserWithPassword`.
// EN: Data type `UserWithPassword`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РѕРїРёСЃС‹РІР°РµС‚ СЃС‚СЂСѓРєС‚СѓСЂСѓ РґР°РЅРЅС‹С… `UserWithPassword`, РєРѕС‚РѕСЂР°СЏ СѓС‡Р°СЃС‚РІСѓРµС‚ РІ Р±РёР·РЅРµСЃ-Р»РѕРіРёРєРµ, API РёР»Рё С‚РµСЃС‚Р°С….
// EN: What it does: UserWithPassword extends a user payload with a plain-text password for creation workflows only.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РєР°Рє РєРѕРЅС‚СЂР°РєС‚ РёР»Рё РѕРїРѕСЂРЅР°СЏ С‚РѕС‡РєР° РґР»СЏ РґСЂСѓРіРёС… С‡Р°СЃС‚РµР№ РїСЂРѕРµРєС‚Р°; РёР·РјРµРЅРµРЅРёСЏ Р·РґРµСЃСЊ С‡Р°СЃС‚Рѕ С‚СЂРµР±СѓСЋС‚ РѕСЃС‚РѕСЂРѕР¶РЅРѕСЃС‚Рё.
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

// RU: РўРёРї РґР°РЅРЅС‹С… `LoginRequest`.
// EN: Data type `LoginRequest`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РѕРїРёСЃС‹РІР°РµС‚ СЃС‚СЂСѓРєС‚СѓСЂСѓ РґР°РЅРЅС‹С… `LoginRequest`, РєРѕС‚РѕСЂР°СЏ СѓС‡Р°СЃС‚РІСѓРµС‚ РІ Р±РёР·РЅРµСЃ-Р»РѕРіРёРєРµ, API РёР»Рё С‚РµСЃС‚Р°С….
// EN: What it does: LoginRequest carries credentials from the login form to the backend session logic.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РєР°Рє РєРѕРЅС‚СЂР°РєС‚ РёР»Рё РѕРїРѕСЂРЅР°СЏ С‚РѕС‡РєР° РґР»СЏ РґСЂСѓРіРёС… С‡Р°СЃС‚РµР№ РїСЂРѕРµРєС‚Р°; РёР·РјРµРЅРµРЅРёСЏ Р·РґРµСЃСЊ С‡Р°СЃС‚Рѕ С‚СЂРµР±СѓСЋС‚ РѕСЃС‚РѕСЂРѕР¶РЅРѕСЃС‚Рё.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RU: РўРёРї РґР°РЅРЅС‹С… `SessionState`.
// EN: Data type `SessionState`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РѕРїРёСЃС‹РІР°РµС‚ СЃС‚СЂСѓРєС‚СѓСЂСѓ РґР°РЅРЅС‹С… `SessionState`, РєРѕС‚РѕСЂР°СЏ СѓС‡Р°СЃС‚РІСѓРµС‚ РІ Р±РёР·РЅРµСЃ-Р»РѕРіРёРєРµ, API РёР»Рё С‚РµСЃС‚Р°С….
// EN: What it does: SessionState is the frontend-facing snapshot of the current authenticated user and their capabilities.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РєР°Рє РєРѕРЅС‚СЂР°РєС‚ РёР»Рё РѕРїРѕСЂРЅР°СЏ С‚РѕС‡РєР° РґР»СЏ РґСЂСѓРіРёС… С‡Р°СЃС‚РµР№ РїСЂРѕРµРєС‚Р°; РёР·РјРµРЅРµРЅРёСЏ Р·РґРµСЃСЊ С‡Р°СЃС‚Рѕ С‚СЂРµР±СѓСЋС‚ РѕСЃС‚РѕСЂРѕР¶РЅРѕСЃС‚Рё.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type SessionState struct {
	Authenticated bool   `json:"authenticated"`
	User          *User  `json:"user,omitempty"`
	CanManage     bool   `json:"canManage"`
	CanAdmin      bool   `json:"canAdmin"`
	CanModerate   bool   `json:"canModerate"`
	Message       string `json:"message,omitempty"`
}

// RU: РўРёРї РґР°РЅРЅС‹С… `CalculationRequest`.
// EN: Data type `CalculationRequest`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РѕРїРёСЃС‹РІР°РµС‚ СЃС‚СЂСѓРєС‚СѓСЂСѓ РґР°РЅРЅС‹С… `CalculationRequest`, РєРѕС‚РѕСЂР°СЏ СѓС‡Р°СЃС‚РІСѓРµС‚ РІ Р±РёР·РЅРµСЃ-Р»РѕРіРёРєРµ, API РёР»Рё С‚РµСЃС‚Р°С….
// EN: What it does: CalculationRequest contains the target amount and per-service weights used to build a calculation.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РєР°Рє РєРѕРЅС‚СЂР°РєС‚ РёР»Рё РѕРїРѕСЂРЅР°СЏ С‚РѕС‡РєР° РґР»СЏ РґСЂСѓРіРёС… С‡Р°СЃС‚РµР№ РїСЂРѕРµРєС‚Р°; РёР·РјРµРЅРµРЅРёСЏ Р·РґРµСЃСЊ С‡Р°СЃС‚Рѕ С‚СЂРµР±СѓСЋС‚ РѕСЃС‚РѕСЂРѕР¶РЅРѕСЃС‚Рё.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type CalculationRequest struct {
	TargetAmount int            `json:"targetAmount"`
	Weights      map[string]int `json:"weights"`
}

// RU: РўРёРї РґР°РЅРЅС‹С… `CalculationItem`.
// EN: Data type `CalculationItem`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РѕРїРёСЃС‹РІР°РµС‚ СЃС‚СЂСѓРєС‚СѓСЂСѓ РґР°РЅРЅС‹С… `CalculationItem`, РєРѕС‚РѕСЂР°СЏ СѓС‡Р°СЃС‚РІСѓРµС‚ РІ Р±РёР·РЅРµСЃ-Р»РѕРіРёРєРµ, API РёР»Рё С‚РµСЃС‚Р°С….
// EN: What it does: CalculationItem is one line of the generated or archived calculation with both business and UI-oriented fields.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РєР°Рє РєРѕРЅС‚СЂР°РєС‚ РёР»Рё РѕРїРѕСЂРЅР°СЏ С‚РѕС‡РєР° РґР»СЏ РґСЂСѓРіРёС… С‡Р°СЃС‚РµР№ РїСЂРѕРµРєС‚Р°; РёР·РјРµРЅРµРЅРёСЏ Р·РґРµСЃСЊ С‡Р°СЃС‚Рѕ С‚СЂРµР±СѓСЋС‚ РѕСЃС‚РѕСЂРѕР¶РЅРѕСЃС‚Рё.
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

// RU: РўРёРї РґР°РЅРЅС‹С… `CalculationResult`.
// EN: Data type `CalculationResult`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РѕРїРёСЃС‹РІР°РµС‚ СЃС‚СЂСѓРєС‚СѓСЂСѓ РґР°РЅРЅС‹С… `CalculationResult`, РєРѕС‚РѕСЂР°СЏ СѓС‡Р°СЃС‚РІСѓРµС‚ РІ Р±РёР·РЅРµСЃ-Р»РѕРіРёРєРµ, API РёР»Рё С‚РµСЃС‚Р°С….
// EN: What it does: CalculationResult is the full response returned after a calculation attempt, including totals and metadata.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РєР°Рє РєРѕРЅС‚СЂР°РєС‚ РёР»Рё РѕРїРѕСЂРЅР°СЏ С‚РѕС‡РєР° РґР»СЏ РґСЂСѓРіРёС… С‡Р°СЃС‚РµР№ РїСЂРѕРµРєС‚Р°; РёР·РјРµРЅРµРЅРёСЏ Р·РґРµСЃСЊ С‡Р°СЃС‚Рѕ С‚СЂРµР±СѓСЋС‚ РѕСЃС‚РѕСЂРѕР¶РЅРѕСЃС‚Рё.
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

// RU: РўРёРї РґР°РЅРЅС‹С… `SaveCalculationRequest`.
// EN: Data type `SaveCalculationRequest`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РѕРїРёСЃС‹РІР°РµС‚ СЃС‚СЂСѓРєС‚СѓСЂСѓ РґР°РЅРЅС‹С… `SaveCalculationRequest`, РєРѕС‚РѕСЂР°СЏ СѓС‡Р°СЃС‚РІСѓРµС‚ РІ Р±РёР·РЅРµСЃ-Р»РѕРіРёРєРµ, API РёР»Рё С‚РµСЃС‚Р°С….
// EN: What it does: SaveCalculationRequest is sent when the frontend persists the current calculation into the archive.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РєР°Рє РєРѕРЅС‚СЂР°РєС‚ РёР»Рё РѕРїРѕСЂРЅР°СЏ С‚РѕС‡РєР° РґР»СЏ РґСЂСѓРіРёС… С‡Р°СЃС‚РµР№ РїСЂРѕРµРєС‚Р°; РёР·РјРµРЅРµРЅРёСЏ Р·РґРµСЃСЊ С‡Р°СЃС‚Рѕ С‚СЂРµР±СѓСЋС‚ РѕСЃС‚РѕСЂРѕР¶РЅРѕСЃС‚Рё.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type SaveCalculationRequest struct {
	Title        string            `json:"title"`
	TargetAmount int               `json:"targetAmount"`
	Items        []CalculationItem `json:"items"`
}

// RU: РўРёРї РґР°РЅРЅС‹С… `SavedCalculation`.
// EN: Data type `SavedCalculation`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РѕРїРёСЃС‹РІР°РµС‚ СЃС‚СЂСѓРєС‚СѓСЂСѓ РґР°РЅРЅС‹С… `SavedCalculation`, РєРѕС‚РѕСЂР°СЏ СѓС‡Р°СЃС‚РІСѓРµС‚ РІ Р±РёР·РЅРµСЃ-Р»РѕРіРёРєРµ, API РёР»Рё С‚РµСЃС‚Р°С….
// EN: What it does: SavedCalculation represents one archived calculation entry as stored in SQLite and shown in history.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РєР°Рє РєРѕРЅС‚СЂР°РєС‚ РёР»Рё РѕРїРѕСЂРЅР°СЏ С‚РѕС‡РєР° РґР»СЏ РґСЂСѓРіРёС… С‡Р°СЃС‚РµР№ РїСЂРѕРµРєС‚Р°; РёР·РјРµРЅРµРЅРёСЏ Р·РґРµСЃСЊ С‡Р°СЃС‚Рѕ С‚СЂРµР±СѓСЋС‚ РѕСЃС‚РѕСЂРѕР¶РЅРѕСЃС‚Рё.
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

// RU: РўРёРї РґР°РЅРЅС‹С… `CopyArchiveServicesResult`.
// EN: Data type `CopyArchiveServicesResult`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РѕРїРёСЃС‹РІР°РµС‚ СЂРµР·СѓР»СЊС‚Р°С‚ РєРѕРїРёСЂРѕРІР°РЅРёСЏ СѓСЃР»СѓРі РёР· Р°СЂС…РёРІРЅРѕРіРѕ СЂР°СЃС‡С‘С‚Р° РІ СЃРїРёСЃРѕРє СѓСЃР»СѓРі Р°РґРјРёРЅРёСЃС‚СЂР°С‚РѕСЂР°.
// EN: What it does: CopyArchiveServicesResult reports how many services were created or updated after importing from an archived calculation.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РёСЃРїРѕР»СЊР·СѓРµС‚СЃСЏ С‚РѕР»СЊРєРѕ РІ Р°РґРјРёРЅСЃРєРѕРј СЃС†РµРЅР°СЂРёРё; РїРѕРјРѕРіР°РµС‚ С„СЂРѕРЅС‚РµРЅРґСѓ РїРѕРєР°Р·Р°С‚СЊ РїРѕРЅСЏС‚РЅРѕРµ СЃРѕРѕР±С‰РµРЅРёРµ РїРѕСЃР»Рµ РёРјРїРѕСЂС‚Р°; РѕС‚РґРµР»СЏРµС‚ СЃС‚Р°С‚РёСЃС‚РёРєСѓ РѕРїРµСЂР°С†РёРё РѕС‚ РїРѕР»РЅРѕРіРѕ СЃРїРёСЃРєР° СѓСЃР»СѓРі.
// EN: Key points: used only in the admin-only archive import flow; lets the frontend show a concise status message; keeps operation stats separate from the full services list.
type CopyArchiveServicesResult struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
}

// RU: РўРёРї РґР°РЅРЅС‹С… `AppBootstrap`.
// EN: Data type `AppBootstrap`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РѕРїРёСЃС‹РІР°РµС‚ СЃС‚СЂСѓРєС‚СѓСЂСѓ РґР°РЅРЅС‹С… `AppBootstrap`, РєРѕС‚РѕСЂР°СЏ СѓС‡Р°СЃС‚РІСѓРµС‚ РІ Р±РёР·РЅРµСЃ-Р»РѕРіРёРєРµ, API РёР»Рё С‚РµСЃС‚Р°С….
// EN: What it does: AppBootstrap aggregates all initial data the frontend needs after startup or refresh.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РєР°Рє РєРѕРЅС‚СЂР°РєС‚ РёР»Рё РѕРїРѕСЂРЅР°СЏ С‚РѕС‡РєР° РґР»СЏ РґСЂСѓРіРёС… С‡Р°СЃС‚РµР№ РїСЂРѕРµРєС‚Р°; РёР·РјРµРЅРµРЅРёСЏ Р·РґРµСЃСЊ С‡Р°СЃС‚Рѕ С‚СЂРµР±СѓСЋС‚ РѕСЃС‚РѕСЂРѕР¶РЅРѕСЃС‚Рё.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type AppBootstrap struct {
	Session             SessionState       `json:"session"`
	Services            []Service          `json:"services"`
	Users               []User             `json:"users"`
	SavedCalculations   []SavedCalculation `json:"savedCalculations"`
	DefaultGroupPercent map[string]float64 `json:"defaultGroupPercent"`
}

// RU: РўРёРї РґР°РЅРЅС‹С… `allocationState`.
// EN: Data type `allocationState`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РѕРїРёСЃС‹РІР°РµС‚ СЃС‚СЂСѓРєС‚СѓСЂСѓ РґР°РЅРЅС‹С… `allocationState`, РєРѕС‚РѕСЂР°СЏ СѓС‡Р°СЃС‚РІСѓРµС‚ РІ Р±РёР·РЅРµСЃ-Р»РѕРіРёРєРµ, API РёР»Рё С‚РµСЃС‚Р°С….
// EN: What it does: allocationState is an internal DP cell used while searching for a good quantity distribution.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РєР°Рє РєРѕРЅС‚СЂР°РєС‚ РёР»Рё РѕРїРѕСЂРЅР°СЏ С‚РѕС‡РєР° РґР»СЏ РґСЂСѓРіРёС… С‡Р°СЃС‚РµР№ РїСЂРѕРµРєС‚Р°; РёР·РјРµРЅРµРЅРёСЏ Р·РґРµСЃСЊ С‡Р°СЃС‚Рѕ С‚СЂРµР±СѓСЋС‚ РѕСЃС‚РѕСЂРѕР¶РЅРѕСЃС‚Рё.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type allocationState struct {
	distance int
	score    int
	count    int
	prev     int
	idx      int
	ok       bool
}

// RU: РўРёРї РґР°РЅРЅС‹С… `groupAllocation`.
// EN: Data type `groupAllocation`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РѕРїРёСЃС‹РІР°РµС‚ СЃС‚СЂСѓРєС‚СѓСЂСѓ РґР°РЅРЅС‹С… `groupAllocation`, РєРѕС‚РѕСЂР°СЏ СѓС‡Р°СЃС‚РІСѓРµС‚ РІ Р±РёР·РЅРµСЃ-Р»РѕРіРёРєРµ, API РёР»Рё С‚РµСЃС‚Р°С….
// EN: What it does: groupAllocation stores the best per-group exact allocation candidate found during structured solving.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РєР°Рє РєРѕРЅС‚СЂР°РєС‚ РёР»Рё РѕРїРѕСЂРЅР°СЏ С‚РѕС‡РєР° РґР»СЏ РґСЂСѓРіРёС… С‡Р°СЃС‚РµР№ РїСЂРѕРµРєС‚Р°; РёР·РјРµРЅРµРЅРёСЏ Р·РґРµСЃСЊ С‡Р°СЃС‚Рѕ С‚СЂРµР±СѓСЋС‚ РѕСЃС‚РѕСЂРѕР¶РЅРѕСЃС‚Рё.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type groupAllocation struct {
	amount     int
	score      int
	quantities []int
	ok         bool
}

// RU: РўРёРї РґР°РЅРЅС‹С… `App`.
// EN: Data type `App`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РѕРїРёСЃС‹РІР°РµС‚ СЃС‚СЂСѓРєС‚СѓСЂСѓ РґР°РЅРЅС‹С… `App`, РєРѕС‚РѕСЂР°СЏ СѓС‡Р°СЃС‚РІСѓРµС‚ РІ Р±РёР·РЅРµСЃ-Р»РѕРіРёРєРµ, API РёР»Рё С‚РµСЃС‚Р°С….
// EN: What it does: App owns application state, the database handle and the current in-memory session.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РєР°Рє РєРѕРЅС‚СЂР°РєС‚ РёР»Рё РѕРїРѕСЂРЅР°СЏ С‚РѕС‡РєР° РґР»СЏ РґСЂСѓРіРёС… С‡Р°СЃС‚РµР№ РїСЂРѕРµРєС‚Р°; РёР·РјРµРЅРµРЅРёСЏ Р·РґРµСЃСЊ С‡Р°СЃС‚Рѕ С‚СЂРµР±СѓСЋС‚ РѕСЃС‚РѕСЂРѕР¶РЅРѕСЃС‚Рё.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type App struct {
	ctx            context.Context
	db             *sql.DB
	mu             sync.RWMutex
	currentSession *User
}

// RU: Р¤СѓРЅРєС†РёСЏ `NewApp`.
// EN: Function `NewApp`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РІС‹РїРѕР»РЅСЏРµС‚ РІСЃРїРѕРјРѕРіР°С‚РµР»СЊРЅРѕРµ РїСЂРµРѕР±СЂР°Р·РѕРІР°РЅРёРµ, РїСЂРѕРІРµСЂРєСѓ РёР»Рё РїРѕРґРіРѕС‚РѕРІРєСѓ РґР°РЅРЅС‹С….
// EN: What it does: NewApp prepares the application object, resolves the working database file and initializes schema/data.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РґР»СЏ СѓСЃС‚РѕР№С‡РёРІРѕСЃС‚Рё Р»РѕРіРёРєРё; РјРѕР¶РµС‚ РёСЃРїРѕР»СЊР·РѕРІР°С‚СЊСЃСЏ СЃСЂР°Р·Сѓ РІ РЅРµСЃРєРѕР»СЊРєРёС… РјРµСЃС‚Р°С…; РёР·РјРµРЅРµРЅРёСЏ СЃС‚РѕРёС‚ РґРµР»Р°С‚СЊ РѕСЃРѕР·РЅР°РЅРЅРѕ.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
