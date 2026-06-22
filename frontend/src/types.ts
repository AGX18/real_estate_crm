export type Role = 'user' | 'admin'
export type LeadStatus = 'Follow_Up' | 'qualified' | 'closed' | 'unqualified'
export type PropertyStatus = 'available' | 'sold' | 'rented'
export type CallOutcome = 'follow_up' | 'qualified' | 'closed' | 'unqualified' | 'no_answer'
export type CallSentiment = 'positive' | 'negative' | 'neutral'
export type AppointmentStatus = 'scheduled' | 'completed' | 'canceled' | 'no_show'

export type Broker = {
  id: number
  tenant_id: string
  username: string
  email: string
  role: Role
  created_at?: string
  updated_at?: string
}

export type LoginResult = {
  token: string
  broker: Broker
}

export type Session = LoginResult & {
  tenantName: string
}

export type Lead = {
  id: number
  tenant_id: string
  phone: string
  description?: NullableText | string | null
  status?: NullableValue<LeadStatus> | LeadStatus | null
  created_at?: string
  updated_at?: string
}

export type Property = {
  id: number
  tenant_id: string
  description?: NullableText | string | null
  price?: unknown
  location?: NullableText | string | null
  area_sqm?: unknown
  type?: NullableValue<string> | string | null
  city?: NullableText | string | null
  governorate?: NullableText | string | null
  bedrooms: number
  bathrooms: number
  status?: NullableValue<PropertyStatus> | PropertyStatus | null
  created_at?: string
  updated_at?: string
}

export type Call = {
  id: number
  tenant_id: string
  lead_id?: NullableNumber | number | null
  transcript?: NullableText | string | null
  details?: NullableText | string | null
  summary?: NullableText | string | null
  sentiment?: NullableValue<CallSentiment> | CallSentiment | null
  outcome?: NullableValue<CallOutcome> | CallOutcome | null
  duration_secs?: NullableNumber | number | null
  created_at?: string
}

export type Appointment = {
  id: number
  tenant_id: string
  lead_id: number
  title: string
  notes?: NullableText | string | null
  status?: NullableValue<AppointmentStatus> | AppointmentStatus | null
  appointment_date?: unknown
  appointment_day?: string
  appointment_time?: unknown
  created_at?: unknown
  updated_at?: unknown
}

export type NullableText = {
  String?: string
  string?: string
  Valid?: boolean
  valid?: boolean
}

export type NullableNumber = {
  Int32?: number
  Int64?: number
  int32?: number
  int64?: number
  Valid?: boolean
  valid?: boolean
}

export type NullableDate = {
  Time?: string
  time?: string
  Valid?: boolean
  valid?: boolean
}

export type NullableTime = {
  Microseconds?: number
  microseconds?: number
  Valid?: boolean
  valid?: boolean
}

export type NullableValue<T extends string> = {
  [key: string]: T | boolean | undefined
  Valid?: boolean
  valid?: boolean
}

export type LoadState = 'idle' | 'loading' | 'ready' | 'error'
export type Page = 'dashboard' | 'leads' | 'appointments' | 'properties' | 'calls' | 'brokers'
export type DashboardMetrics = {
  qualified: number
  followUps: number
  available: number
  closedInventory: number
  qualifiedCalls: number
  upcomingAppointments: number
}
