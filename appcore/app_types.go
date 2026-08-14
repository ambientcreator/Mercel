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
	RoleAdmin:                     {Department: DepartmentGlobal, Level: 5, Label: "Администратор"},
	RoleGlobalDirector:            {Department: DepartmentGlobal, Level: 4, Label: "Генеральный директор"},
	RoleExecutiveDirector:         {Department: DepartmentGlobal, Level: 4, Label: "Исполнительный директор"},
	RoleTechnicalDirector:         {Department: DepartmentGlobal, Level: 4, Label: "Технический директор"},
	RoleSupportHead:               {Department: DepartmentSupport, Level: 3, Label: "Начальник отдела технической поддержки"},
	RoleSupportSysadmin:           {Department: DepartmentSupport, Level: 3, Label: "Системный администратор"},
	RoleSupportSenior:             {Department: DepartmentSupport, Level: 2, Label: "Старший специалист технической поддержки"},
	RoleSupportEmployee:           {Department: DepartmentSupport, Level: 1, Label: "Специалист технической поддержки"},
	RoleTechnicalHead:             {Department: DepartmentTechnical, Level: 3, Label: "Начальник технического отдела"},
	RoleTechnicalSenior:           {Department: DepartmentTechnical, Level: 2, Label: "Старший техник"},
	RoleTechnicalEmployee:         {Department: DepartmentTechnical, Level: 1, Label: "Техник"},
	RoleTelecomDirector:           {Department: DepartmentTelecom, Level: 3, Label: "Директор по строительству сетей связи"},
	RoleTelecomHead:               {Department: DepartmentTelecom, Level: 3, Label: "Начальник отдела строительства сетей связи"},
	RoleTelecomSeniorVOLS:         {Department: DepartmentTelecom, Level: 2, Label: "Старший инженер ВОЛС"},
	RoleTelecomSeniorLVS:          {Department: DepartmentTelecom, Level: 2, Label: "Старший инженер ЛВС"},
	RoleTelecomEmployeeVOLS:       {Department: DepartmentTelecom, Level: 1, Label: "Инженер ВОЛС"},
	RoleTelecomEmployeeLVS:        {Department: DepartmentTelecom, Level: 1, Label: "Инженер ЛВС"},
	RoleSKUDHead:                  {Department: DepartmentSKUD, Level: 3, Label: "Начальник отдела технического обслуживания СКУД"},
	RoleSKUDProjectManager:        {Department: DepartmentSKUD, Level: 2, Label: "Менеджер проектов СКУД"},
	RoleSKUDSeniorService:         {Department: DepartmentSKUD, Level: 2, Label: "Старший сервисный инженер СКУД"},
	RoleSKUDSeniorInstaller:       {Department: DepartmentSKUD, Level: 2, Label: "Старший монтажник СКУД"},
	RoleSKUDServiceEngineer:       {Department: DepartmentSKUD, Level: 1, Label: "Сервисный инженер СКУД"},
	RoleSKUDInstaller:             {Department: DepartmentSKUD, Level: 1, Label: "Монтажник СКУД"},
	RoleApprovalHead:              {Department: DepartmentApproval, Level: 3, Label: "Начальник отдела согласований"},
	RoleApprovalSenior:            {Department: DepartmentApproval, Level: 2, Label: "Старший менеджер согласований"},
	RoleApprovalEmployee:          {Department: DepartmentApproval, Level: 1, Label: "Менеджер согласований"},
	RoleMarketingHead:             {Department: DepartmentMarketing, Level: 3, Label: "Начальник отдела рекламы и маркетинга"},
	RoleMarketingCourier:          {Department: DepartmentMarketing, Level: 1, Label: "Курьер"},
	RoleCommercialDirector:        {Department: DepartmentCommercial, Level: 4, Label: "Коммерческий директор"},
	RoleCommercialSubscriberHead:  {Department: DepartmentCommercial, Level: 3, Label: "Начальник абонентского отдела"},
	RoleCommercialActiveSalesHead: {Department: DepartmentCommercial, Level: 3, Label: "Менеджер активных продаж"},
	RoleCommercialSeniorMRK:       {Department: DepartmentCommercial, Level: 2, Label: "Старший менеджер МРК"},
	RoleCommercialSeniorMRYU:      {Department: DepartmentCommercial, Level: 2, Label: "Старший менеджер МРЮ"},
	RoleCommercialEmployeeMRK:     {Department: DepartmentCommercial, Level: 1, Label: "Менеджер МРК"},
	RoleCommercialEmployeeMRYU:    {Department: DepartmentCommercial, Level: 1, Label: "Менеджер МРЮ"},
	RoleFinanceHead:               {Department: DepartmentFinance, Level: 3, Label: "Главный бухгалтер"},
	RoleFinanceEmployee:           {Department: DepartmentFinance, Level: 1, Label: "Помощник бухгалтера"},
	RoleLegalEmployee:             {Department: DepartmentLegal, Level: 1, Label: "Юрист"},
	RoleDevelopmentHead:           {Department: DepartmentDevelopment, Level: 3, Label: "Начальник отдела развития"},
	RoleDevelopmentSenior:         {Department: DepartmentDevelopment, Level: 2, Label: "Старший менеджер развития"},
	RoleDevelopmentEmployee:       {Department: DepartmentDevelopment, Level: 1, Label: "Менеджер развития"},
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

// EN: Data type `Service`.
//
// EN: What it does: Service describes one billable position available to the current user: pricing, unit, ownership and allocation settings.
//
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

// EN: Data type `UpsertServiceRequest`.
//
// EN: What it does: UpsertServiceRequest is the payload used by the frontend when a user creates or edits a service definition.
//
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type UpsertServiceRequest struct {
	ID                int64    `json:"id"`
	Name              string   `json:"name"`
	Unit              string   `json:"unit"`
	Rate              int      `json:"rate"`
	Category          string   `json:"category"`
	AllocationPercent *float64 `json:"allocationPercent"`
}

// EN: Data type `User`.
//
// EN: What it does: User stores public account information that can be safely returned to the frontend without password data.
//
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
	ActTemplate           string `json:"actTemplate"`
}

// EN: Data type `UserWithPassword`.
//
// EN: What it does: UserWithPassword extends a user payload with a plain-text password for creation workflows only.
//
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

// EN: Data type `LoginRequest`.
//
// EN: What it does: LoginRequest carries credentials from the login form to the backend session logic.
//
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// EN: Data type `SessionState`.
//
// EN: What it does: SessionState is the frontend-facing snapshot of the current authenticated user and their capabilities.
//
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type SessionState struct {
	Authenticated bool   `json:"authenticated"`
	User          *User  `json:"user,omitempty"`
	CanManage     bool   `json:"canManage"`
	CanAdmin      bool   `json:"canAdmin"`
	CanModerate   bool   `json:"canModerate"`
	Message       string `json:"message,omitempty"`
}

// EN: Data type `CalculationRequest`.
//
// EN: What it does: CalculationRequest contains the target amount and per-service weights used to build a calculation.
//
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type CalculationRequest struct {
	TargetAmount int            `json:"targetAmount"`
	Weights      map[string]int `json:"weights"`
}

// EN: Data type `CalculationItem`.
//
// EN: What it does: CalculationItem is one line of the generated or archived calculation with both business and UI-oriented fields.
//
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

// EN: Data type `CalculationResult`.
//
// EN: What it does: CalculationResult is the full response returned after a calculation attempt, including totals and metadata.
//
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

// EN: Data type `SaveCalculationRequest`.
//
// EN: What it does: SaveCalculationRequest is sent when the frontend persists the current calculation into the archive.
//
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type SaveCalculationRequest struct {
	Title        string            `json:"title"`
	TargetAmount int               `json:"targetAmount"`
	Items        []CalculationItem `json:"items"`
}

type UpdateSavedCalculationRequest struct {
	ID    int64             `json:"id"`
	Items []CalculationItem `json:"items"`
}

// EN: Data type `SavedCalculation`.
//
// EN: What it does: SavedCalculation represents one archived calculation entry as stored in SQLite and shown in history.
//
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

// EN: Data type `CopyArchiveServicesResult`.
//
// EN: What it does: CopyArchiveServicesResult reports how many services were created or updated after importing from an archived calculation.
//
// EN: Key points: used only in the admin-only archive import flow; lets the frontend show a concise status message; keeps operation stats separate from the full services list.
type CopyArchiveServicesResult struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
}

// EN: Data type `AppBootstrap`.
//
// EN: What it does: AppBootstrap aggregates all initial data the frontend needs after startup or refresh.
//
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type AppBootstrap struct {
	Session             SessionState       `json:"session"`
	Services            []Service          `json:"services"`
	Users               []User             `json:"users"`
	SavedCalculations   []SavedCalculation `json:"savedCalculations"`
	DefaultGroupPercent map[string]float64 `json:"defaultGroupPercent"`
}

// EN: Data type `allocationState`.
//
// EN: What it does: allocationState is an internal DP cell used while searching for a good quantity distribution.
//
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type allocationState struct {
	distance int
	score    int
	count    int
	prev     int
	idx      int
	ok       bool
}

// EN: Data type `groupAllocation`.
//
// EN: What it does: groupAllocation stores the best per-group exact allocation candidate found during structured solving.
//
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type groupAllocation struct {
	amount     int
	score      int
	quantities []int
	ok         bool
}

// EN: Data type `App`.
//
// EN: What it does: App owns application state, the database handle and the current in-memory session.
//
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type App struct {
	ctx            context.Context
	db             *sql.DB
	mu             sync.RWMutex
	currentSession *User
}

// EN: Function `NewApp`.
//
// EN: What it does: NewApp prepares the application object, resolves the working database file and initializes schema/data.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
