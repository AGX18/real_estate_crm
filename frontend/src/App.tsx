import { useEffect, useMemo, useState } from 'react'
import type { FormEvent } from 'react'
import heroImage from './assets/hero.png'
import './App.css'

type Role = 'user' | 'admin'
type LeadStatus = 'Follow_Up' | 'qualified' | 'closed' | 'unqualified'
type PropertyStatus = 'available' | 'sold' | 'rented'
type CallOutcome = 'follow_up' | 'qualified' | 'closed' | 'unqualified' | 'no_answer'
type CallSentiment = 'positive' | 'negative' | 'neutral'

type Broker = {
  id: number
  tenant_id: string
  username: string
  email: string
  role: Role
  created_at?: string
  updated_at?: string
}

type LoginResult = {
  token: string
  broker: Broker
}

type Session = LoginResult & {
  tenantName: string
}

type Lead = {
  id: number
  tenant_id: string
  phone: string
  description?: NullableText | string | null
  status?: NullableValue<LeadStatus> | LeadStatus | null
  created_at?: string
  updated_at?: string
}

type Property = {
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

type Call = {
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

type NullableText = {
  String?: string
  string?: string
  Valid?: boolean
  valid?: boolean
}

type NullableNumber = {
  Int32?: number
  Int64?: number
  int32?: number
  int64?: number
  Valid?: boolean
  valid?: boolean
}

type NullableValue<T extends string> = {
  [key: string]: T | boolean | undefined
  Valid?: boolean
  valid?: boolean
}

type LoadState = 'idle' | 'loading' | 'ready' | 'error'

const apiBase = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
const sessionKey = 'real_estate_crm_session'

const leadLabels: Record<LeadStatus, string> = {
  Follow_Up: 'Follow up',
  qualified: 'Qualified',
  closed: 'Closed',
  unqualified: 'Unqualified',
}

const propertyStatusLabels: Record<PropertyStatus, string> = {
  available: 'Available',
  sold: 'Sold',
  rented: 'Rented',
}

const callOutcomeLabels: Record<CallOutcome, string> = {
  follow_up: 'Follow up',
  qualified: 'Qualified',
  closed: 'Closed',
  unqualified: 'Unqualified',
  no_answer: 'No answer',
}

function App() {
  const [session, setSession] = useState<Session | null>(() => loadSession())
  const [leads, setLeads] = useState<Lead[]>([])
  const [properties, setProperties] = useState<Property[]>([])
  const [brokers, setBrokers] = useState<Broker[]>([])
  const [calls, setCalls] = useState<Call[]>([])
  const [loadState, setLoadState] = useState<LoadState>('idle')
  const [error, setError] = useState('')
  const [propertyFilter, setPropertyFilter] = useState<PropertyStatus | 'all'>('all')
  const [selectedLeadID, setSelectedLeadID] = useState<number | null>(null)

  const isAdmin = session?.broker.role === 'admin'
  const refreshDashboard = () => {
    if (!session) {
      return
    }
    reload(session, isAdmin, setLeads, setProperties, setBrokers, setCalls, setLoadState, setError, setSelectedLeadID)
  }

  useEffect(() => {
    if (!session) {
      return
    }

    let cancelled = false
    Promise.all([
      apiGet<Lead[]>(`/tenants/${session.broker.tenant_id}/leads`, session.token),
      apiGet<Property[]>(`/tenants/${session.broker.tenant_id}/properties`, session.token),
      apiGet<Call[]>(`/tenants/${session.broker.tenant_id}/calls`, session.token),
      isAdmin
        ? apiGet<Broker[]>(`/tenants/${session.broker.tenant_id}/brokers`, session.token)
        : Promise.resolve([]),
    ])
      .then(([nextLeads, nextProperties, nextCalls, nextBrokers]) => {
        if (cancelled) {
          return
        }
        applyDashboardData(
          nextLeads,
          nextProperties,
          nextBrokers,
          nextCalls,
          setLeads,
          setProperties,
          setBrokers,
          setCalls,
          setSelectedLeadID,
        )
        setLoadState('ready')
      })
      .catch((err: Error) => {
        if (cancelled) {
          return
        }
        setError(err.message)
        setLoadState('error')
      })

    return () => {
      cancelled = true
    }
  }, [isAdmin, session])

  const metrics = useMemo(() => {
    const qualified = leads.filter((lead) => leadStatus(lead) === 'qualified').length
    const followUps = leads.filter((lead) => leadStatus(lead) === 'Follow_Up').length
    const available = properties.filter((property) => propertyStatus(property) === 'available').length
    const closedInventory = properties.filter((property) => {
      const status = propertyStatus(property)
      return status === 'sold' || status === 'rented'
    }).length
    const qualifiedCalls = calls.filter((call) => callOutcome(call) === 'qualified').length

    return { qualified, followUps, available, closedInventory, qualifiedCalls }
  }, [calls, leads, properties])

  const propertyCounts = useMemo(
    () => ({
      all: properties.length,
      available: properties.filter((property) => propertyStatus(property) === 'available').length,
      sold: properties.filter((property) => propertyStatus(property) === 'sold').length,
      rented: properties.filter((property) => propertyStatus(property) === 'rented').length,
    }),
    [properties],
  )

  const filteredProperties = useMemo(
    () =>
      propertyFilter === 'all'
        ? properties
        : properties.filter((property) => propertyStatus(property) === propertyFilter),
    [properties, propertyFilter],
  )

  const selectedLead = leads.find((lead) => lead.id === selectedLeadID) ?? leads[0]

  function handleLogin(nextSession: Session) {
    localStorage.setItem(sessionKey, JSON.stringify(nextSession))
    setSession(nextSession)
  }

  function handleLogout() {
    localStorage.removeItem(sessionKey)
    setSession(null)
    setLeads([])
    setProperties([])
    setBrokers([])
    setCalls([])
    setSelectedLeadID(null)
    setLoadState('idle')
  }

  if (!session) {
    return <LoginScreen onLogin={handleLogin} />
  }

  return (
    <main className="app-shell">
      <aside className="sidebar" aria-label="Main navigation">
        <div className="brand">
          <span className="brand-mark">RE</span>
          <div>
            <strong>{session.tenantName}</strong>
            <span>{session.broker.role} dashboard</span>
          </div>
        </div>

        <nav className="nav-list">
          {['Dashboard', 'Leads', 'Properties', 'Calls', ...(isAdmin ? ['Brokers'] : [])].map((item) => (
            <a className={item === 'Dashboard' ? 'active' : ''} href={`#${item.toLowerCase()}`} key={item}>
              {item}
            </a>
          ))}
        </nav>

        <div className="api-status">
          <span className="status-dot online" />
          <div>
            <strong>{session.broker.username}</strong>
            <span>{session.broker.email}</span>
          </div>
        </div>
      </aside>

      <section className="workspace">
        <header className="topbar">
          <div>
            <p className="eyebrow">{session.tenantName}</p>
            <h1>Real estate command center</h1>
          </div>
          <div className="topbar-actions">
            {isAdmin && <PropertyImport session={session} onImported={refreshDashboard} />}
            <button className="secondary-button" onClick={refreshDashboard}>Refresh</button>
            <button className="secondary-button" onClick={handleLogout}>Logout</button>
          </div>
        </header>

        <section className="hero-band">
          <div className="hero-copy">
            <p className="eyebrow">{isAdmin ? 'Admin workspace' : 'Broker workspace'}</p>
            <h2>Manage tenant-scoped leads and property inventory from one dashboard.</h2>
          </div>
          <img src={heroImage} alt="CRM dashboard preview" />
        </section>

        {loadState === 'error' && <p className="error-banner">{error}</p>}
        {loadState === 'loading' && <p className="loading-banner">Loading tenant data...</p>}

        <section className="metrics-grid" aria-label="Business metrics">
          <Metric label="Qualified leads" value={String(metrics.qualified)} trend="ready to move" />
          <Metric label="Follow-ups" value={String(metrics.followUps)} trend="needs next action" />
          <Metric label="Available units" value={String(metrics.available)} trend="open inventory" />
          <Metric label="Qualified calls" value={String(metrics.qualifiedCalls)} trend="voice outcomes" />
        </section>

        <section className="content-grid">
          <div className="panel pipeline-panel" id="leads">
            <div className="panel-header">
              <div>
                <p className="eyebrow">Lead pipeline</p>
                <h2>Leads</h2>
              </div>
            </div>

            <div className="lead-list">
              {leads.length === 0 && <EmptyState message="No leads found for this tenant." />}
              {leads.map((lead) => {
                const description = leadDescription(lead)
                return (
                  <button
                    className={`lead-card ${selectedLead?.id === lead.id ? 'selected' : ''}`}
                    key={lead.id}
                    onClick={() => setSelectedLeadID(lead.id)}
                  >
                    <span className={`badge ${leadStatusClass(leadStatus(lead))}`}>{leadLabels[leadStatus(lead)]}</span>
                    <strong>{lead.phone}</strong>
                    <span>{description.summary || 'No description'}</span>
                    <small>{dateLabel(lead.created_at)}</small>
                  </button>
                )
              })}
            </div>
          </div>

          <div className="panel detail-panel">
            <div className="panel-header">
              <div>
                <p className="eyebrow">Selected lead</p>
                <h2>{selectedLead ? selectedLead.phone : 'No lead selected'}</h2>
              </div>
              {selectedLead && <span className={`badge ${leadStatusClass(leadStatus(selectedLead))}`}>{leadLabels[leadStatus(selectedLead)]}</span>}
            </div>
            {selectedLead ? (
              <dl className="detail-list lead-detail-list">
                <LeadDescriptionDetails lead={selectedLead} />
                <div>
                  <dt>Created</dt>
                  <dd>{dateLabel(selectedLead.created_at)}</dd>
                </div>
              </dl>
            ) : (
              <EmptyState message="Select a lead to see details." />
            )}
          </div>
        </section>

        <section className="content-grid inventory-grid" id="properties">
          <div className="panel wide-panel">
            <div className="panel-header">
              <div>
                <p className="eyebrow">Property inventory</p>
                <h2>Properties</h2>
              </div>
              <div className="filter-tabs" aria-label="Property filters">
                {(['all', 'available', 'sold', 'rented'] as const).map((status) => (
                  <button
                    className={propertyFilter === status ? 'active' : ''}
                    key={status}
                    onClick={() => setPropertyFilter(status)}
                  >
                    {status} {propertyCounts[status]}
                  </button>
                ))}
              </div>
            </div>

            <div className="property-table">
              {filteredProperties.length === 0 && <EmptyState message="No properties match this filter." />}
              {filteredProperties.map((property) => (
                <article className="property-row" key={property.id}>
                  <div>
                    <strong>{textValue(property.description) || `Property #${property.id}`}</strong>
                    <span>{valueLabel(propertyType(property))} · {textValue(property.city) || 'No city'} · {textValue(property.location) || 'No location'}</span>
                  </div>
                  <div className="property-specs">
                    <span>{property.bedrooms} bed</span>
                    <span>{property.bathrooms} bath</span>
                    <span>{areaLabel(property.area_sqm)}</span>
                    <span>{propertyStatusLabels[propertyStatus(property)]}</span>
                  </div>
                  <strong>{priceLabel(property.price)}</strong>
                </article>
              ))}
            </div>
          </div>

          {isAdmin && (
            <div className="panel" id="brokers">
              <div className="panel-header">
                <div>
                  <p className="eyebrow">Admin only</p>
                  <h2>Brokers</h2>
                </div>
              </div>
              <BrokerCreateForm
                session={session}
                onCreated={refreshDashboard}
              />
              <div className="broker-list">
                {brokers.length === 0 && <EmptyState message="No brokers found for this tenant." />}
                {brokers.map((broker) => (
                  <article className="broker-card" key={broker.id}>
                    <span className="broker-avatar">{broker.username.slice(0, 2).toUpperCase()}</span>
                    <div>
                      <strong>{broker.username}</strong>
                      <small>{broker.email}</small>
                    </div>
                    <span className={`role-badge ${broker.role}`}>{broker.role}</span>
                  </article>
                ))}
              </div>
            </div>
          )}
        </section>

        <section className="panel calls-panel" id="calls">
          <div className="panel-header">
            <div>
              <p className="eyebrow">Voice activity</p>
              <h2>Calls</h2>
            </div>
          </div>
          <div className="calls-list">
            {calls.length === 0 && <EmptyState message="No calls found for this tenant." />}
            {calls.map((call) => (
              <CallCard call={call} leads={leads} key={call.id} />
            ))}
          </div>
        </section>
      </section>
    </main>
  )
}

function LoginScreen({ onLogin }: { onLogin: (session: Session) => void }) {
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [tenantName, setTenantName] = useState('')
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSubmitting(true)
    setError('')

    const path = mode === 'login' ? '/login' : '/register'
    const body =
      mode === 'login'
        ? {
            tenant_name: tenantName.trim(),
            email: email.trim(),
            password,
          }
        : {
            tenant_name: tenantName.trim(),
            admin_username: username.trim(),
            admin_email: email.trim(),
            admin_password: password,
          }

    apiPost<LoginResult>(path, body)
      .then((result) => onLogin({ ...result, tenantName: tenantName.trim() }))
      .catch((err: Error) => setError(err.message))
      .finally(() => setSubmitting(false))
  }

  return (
    <main className="login-shell">
      <form className="login-panel" onSubmit={submit}>
        <div className="brand">
          <span className="brand-mark">RE</span>
          <div>
            <strong>EstateDesk</strong>
            <span>Broker CRM</span>
          </div>
        </div>
        <div>
          <p className="eyebrow">{mode === 'login' ? 'Tenant login' : 'Tenant registration'}</p>
          <h1>{mode === 'login' ? 'Sign in with your company name' : 'Create a tenant and admin account'}</h1>
        </div>
        <div className="mode-switch" aria-label="Authentication mode">
          <button className={mode === 'login' ? 'active' : ''} type="button" onClick={() => setMode('login')}>
            Sign in
          </button>
          <button className={mode === 'register' ? 'active' : ''} type="button" onClick={() => setMode('register')}>
            Register
          </button>
        </div>
        <label>
          Tenant name
          <input value={tenantName} onChange={(event) => setTenantName(event.target.value)} required />
        </label>
        {mode === 'register' && (
          <label>
            Admin username
            <input value={username} onChange={(event) => setUsername(event.target.value)} required />
          </label>
        )}
        <label>
          {mode === 'login' ? 'Email' : 'Admin email'}
          <input type="email" value={email} onChange={(event) => setEmail(event.target.value)} required />
        </label>
        <label>
          {mode === 'login' ? 'Password' : 'Admin password'}
          <input type="password" value={password} onChange={(event) => setPassword(event.target.value)} required />
        </label>
        {error && <p className="form-error">{error}</p>}
        <button className="primary-button" disabled={submitting}>
          {submitting ? 'Working...' : mode === 'login' ? 'Sign in' : 'Create tenant'}
        </button>
      </form>
    </main>
  )
}

function BrokerCreateForm({ session, onCreated }: { session: Session; onCreated: () => void }) {
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState<Role>('user')
  const [submitting, setSubmitting] = useState(false)
  const [message, setMessage] = useState('')

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSubmitting(true)
    setMessage('')

    apiPost<Broker>(
      `/tenants/${session.broker.tenant_id}/brokers`,
      {
        username: username.trim(),
        email: email.trim(),
        password,
        role,
      },
      session.token,
    )
      .then(() => {
        setUsername('')
        setEmail('')
        setPassword('')
        setRole('user')
        setMessage('Broker created.')
        onCreated()
      })
      .catch((err: Error) => setMessage(err.message))
      .finally(() => setSubmitting(false))
  }

  return (
    <form className="broker-form" onSubmit={submit}>
      <label>
        Username
        <input value={username} onChange={(event) => setUsername(event.target.value)} required />
      </label>
      <label>
        Email
        <input type="email" value={email} onChange={(event) => setEmail(event.target.value)} required />
      </label>
      <label>
        Password
        <input type="password" value={password} onChange={(event) => setPassword(event.target.value)} required />
      </label>
      <label>
        Role
        <select value={role} onChange={(event) => setRole(event.target.value as Role)}>
          <option value="user">User</option>
          <option value="admin">Admin</option>
        </select>
      </label>
      <button className="primary-button" disabled={submitting}>
        {submitting ? 'Creating...' : 'Create broker'}
      </button>
      {message && <p className="form-note">{message}</p>}
    </form>
  )
}

function PropertyImport({ session, onImported }: { session: Session; onImported: () => void }) {
  const [file, setFile] = useState<File | null>(null)
  const [message, setMessage] = useState('')
  const [uploading, setUploading] = useState(false)

  function upload() {
    if (!file) {
      setMessage('Choose a JSON file first.')
      return
    }

    const body = new FormData()
    body.append('file', file)
    setUploading(true)
    setMessage('')

    fetch(`${apiBase}/tenants/${session.broker.tenant_id}/properties/import`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${session.token}` },
      body,
    })
      .then(async (response) => {
        if (!response.ok) {
          throw new Error(await responseError(response))
        }
        setMessage('Properties imported.')
        setFile(null)
        onImported()
      })
      .catch((err: Error) => setMessage(err.message))
      .finally(() => setUploading(false))
  }

  return (
    <div className="import-control">
      <input
        aria-label="Property import file"
        type="file"
        accept="application/json,.json"
        onChange={(event) => setFile(event.target.files?.[0] ?? null)}
      />
      <button className="primary-button" onClick={upload} disabled={uploading}>
        {uploading ? 'Uploading...' : 'Import properties'}
      </button>
      {message && <span>{message}</span>}
    </div>
  )
}

function Metric({ label, value, trend }: { label: string; value: string; trend: string }) {
  return (
    <article className="metric-card">
      <span>{label}</span>
      <strong>{value}</strong>
      <small>{trend}</small>
    </article>
  )
}

function EmptyState({ message }: { message: string }) {
  return <p className="empty-state">{message}</p>
}

function LeadDescriptionDetails({ lead }: { lead: Lead }) {
  const description = leadDescription(lead)

  if (!description.summary && !description.intent && description.qualifications.length === 0) {
    return (
      <div>
        <dt>Description</dt>
        <dd>No description</dd>
      </div>
    )
  }

  return (
    <>
      {description.summary && (
        <div>
          <dt>Summary</dt>
          <dd>{description.summary}</dd>
        </div>
      )}
      {description.intent && (
        <div>
          <dt>Intent</dt>
          <dd>{humanizeValue(description.intent)}</dd>
        </div>
      )}
      {description.qualifications.length > 0 && (
        <div>
          <dt>Qualification</dt>
          <dd>
            <div className="qualification-list">
              {description.qualifications.map(([key, value]) => (
                <span key={key}>
                  <b>{humanizeValue(key)}</b>
                  {value}
                </span>
              ))}
            </div>
          </dd>
        </div>
      )}
    </>
  )
}

function CallCard({ call, leads }: { call: Call; leads: Lead[] }) {
  const outcome = callOutcome(call)
  const lead = leads.find((item) => item.id === numberValue(call.lead_id))
  const rawSummary = textValue(call.summary)
  const summary = callSummaryText(rawSummary)
  const details = [...parseCallDetails(textValue(call.details)), ...parseStructuredCallSummary(rawSummary)]
  const transcript = textValue(call.transcript)

  return (
    <article className="call-card">
      <div className="call-card-main">
        <div className="call-card-title">
          <span className={`badge ${callOutcomeClass(outcome)}`}>{callOutcomeLabels[outcome]}</span>
          <strong>{summary || 'No call summary'}</strong>
          <small>{dateLabel(call.created_at)}</small>
        </div>
        <div className="call-meta">
          <span>{callSentiment(call)}</span>
          <span>{durationLabel(call.duration_secs)}</span>
          <span>{lead?.phone ?? 'No linked lead'}</span>
        </div>
      </div>

      {details.length > 0 && (
        <div className="call-detail-grid">
          {details.map(([key, value]) => (
            <div key={key}>
              <dt>{humanizeValue(key)}</dt>
              <dd>{formatCallDetailValue(value)}</dd>
            </div>
          ))}
        </div>
      )}

      {transcript && (
        <details className="transcript-panel">
          <summary>Transcript</summary>
          <pre>{transcript}</pre>
        </details>
      )}
    </article>
  )
}

async function reload(
  session: Session,
  isAdmin: boolean,
  setLeads: (value: Lead[]) => void,
  setProperties: (value: Property[]) => void,
  setBrokers: (value: Broker[]) => void,
  setCalls: (value: Call[]) => void,
  setLoadState: (value: LoadState) => void,
  setError: (value: string) => void,
  setSelectedLeadID?: (value: number | null) => void,
) {
  setLoadState('loading')
  setError('')
  try {
    const [nextLeads, nextProperties, nextCalls, nextBrokers] = await Promise.all([
      apiGet<Lead[]>(`/tenants/${session.broker.tenant_id}/leads`, session.token),
      apiGet<Property[]>(`/tenants/${session.broker.tenant_id}/properties`, session.token),
      apiGet<Call[]>(`/tenants/${session.broker.tenant_id}/calls`, session.token),
      isAdmin
        ? apiGet<Broker[]>(`/tenants/${session.broker.tenant_id}/brokers`, session.token)
        : Promise.resolve([]),
    ])
    applyDashboardData(
      nextLeads,
      nextProperties,
      nextBrokers,
      nextCalls,
      setLeads,
      setProperties,
      setBrokers,
      setCalls,
      setSelectedLeadID,
    )
    setLoadState('ready')
  } catch (err) {
    setError(err instanceof Error ? err.message : 'Failed to reload data')
    setLoadState('error')
  }
}

function applyDashboardData(
  nextLeads: Lead[] | null | undefined,
  nextProperties: Property[] | null | undefined,
  nextBrokers: Broker[] | null | undefined,
  nextCalls: Call[] | null | undefined,
  setLeads: (value: Lead[]) => void,
  setProperties: (value: Property[]) => void,
  setBrokers: (value: Broker[]) => void,
  setCalls: (value: Call[]) => void,
  setSelectedLeadID?: (value: number | null) => void,
) {
  const safeLeads = asArray(nextLeads)
  setLeads(safeLeads)
  setProperties(asArray(nextProperties))
  setBrokers(asArray(nextBrokers))
  setCalls(asArray(nextCalls))
  setSelectedLeadID?.(safeLeads[0]?.id ?? null)
}

function asArray<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : []
}

async function apiGet<T>(path: string, token: string): Promise<T> {
  const response = await fetch(`${apiBase}${path}`, {
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!response.ok) {
    throw new Error(await responseError(response))
  }
  return response.json() as Promise<T>
}

async function apiPost<T>(path: string, body: unknown, token?: string): Promise<T> {
  const response = await fetch(`${apiBase}${path}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify(body),
  })
  if (!response.ok) {
    throw new Error(await responseError(response))
  }
  return response.json() as Promise<T>
}

async function responseError(response: Response) {
  const fallback = `Request failed with ${response.status}`
  try {
    const body = await response.json()
    return body.error ?? body.message ?? fallback
  } catch {
    return fallback
  }
}

function loadSession(): Session | null {
  const stored = localStorage.getItem(sessionKey)
  if (!stored) {
    return null
  }
  try {
    return JSON.parse(stored) as Session
  } catch {
    localStorage.removeItem(sessionKey)
    return null
  }
}

function textValue(value: NullableText | string | null | undefined) {
  if (!value) {
    return ''
  }
  if (typeof value === 'string') {
    return value
  }
  const valid = value.Valid ?? value.valid ?? true
  return valid ? value.String ?? value.string ?? '' : ''
}

function leadDescription(lead: Lead) {
  return parseLeadDescription(textValue(lead.description))
}

function parseLeadDescription(value: string): {
  summary: string
  intent: string
  qualifications: Array<[string, string]>
} {
  const lines = value
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)

  const qualificationLine = lines.find((line) => line.startsWith('Qualification:'))
  const intentLine = lines.find((line) => line.startsWith('Intent:'))
  const summaryLines = lines.filter(
    (line) => !line.startsWith('Qualification:') && !line.startsWith('Intent:'),
  )

  return {
    summary: summaryLines.join(' ').trim() || value.trim(),
    intent: intentLine?.replace('Intent:', '').trim() ?? '',
    qualifications: parseQualification(qualificationLine?.replace('Qualification:', '').trim() ?? ''),
  }
}

function parseQualification(value: string): Array<[string, string]> {
  if (!value) {
    return []
  }

  try {
    const parsed = JSON.parse(value) as Record<string, unknown>
    return Object.entries(parsed)
      .filter(([, item]) => item !== null && item !== undefined && item !== '')
      .map(([key, item]) => [key, String(item)])
  } catch {
    return [['details', value]]
  }
}

function parseCallDetails(value: string): Array<[string, string]> {
  return value
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => {
      const separatorIndex = line.indexOf(':')
      if (separatorIndex === -1) {
        return ['details', line] as [string, string]
      }
      return [line.slice(0, separatorIndex).trim(), line.slice(separatorIndex + 1).trim()] as [string, string]
    })
    .filter(([, item]) => item !== '')
}

function parseStructuredCallSummary(value: string): Array<[string, string]> {
  const labels = /(phone number|phone|budget|rooms|location|property type|sentiment|call outcome)\s*:/gi
  const matches = [...value.matchAll(labels)]
  if (matches.length === 0) {
    return []
  }

  return matches
    .map((match, index) => {
      const start = (match.index ?? 0) + match[0].length
      const end = matches[index + 1]?.index ?? value.length
      return [normalizeDetailKey(match[1]), value.slice(start, end).trim()] as [string, string]
    })
    .filter(([, item]) => item !== '' && !item.toLowerCase().startsWith('assistant:'))
}

function callSummaryText(value: string) {
  const fields = Object.fromEntries(parseStructuredCallSummary(value))
  if (Object.keys(fields).length === 0) {
    return value.replace(/\s*Sentiment\s+(positive|negative|neutral)\.\s*$/i, '').trim()
  }

  const outcome = fields.call_outcome ? humanizeValue(fields.call_outcome) : 'Call'
  const phone = fields.phone_number ?? fields.phone
  return phone ? `${outcome} from ${phone}` : outcome
}

function normalizeDetailKey(value: string) {
  return value.trim().toLowerCase().replaceAll(' ', '_')
}

function formatCallDetailValue(value: string) {
  if (!value.startsWith('{')) {
    return humanizeValue(value)
  }

  try {
    const parsed = JSON.parse(value) as Record<string, unknown>
    return Object.entries(parsed)
      .filter(([, item]) => item !== null && item !== undefined && item !== '')
      .map(([key, item]) => `${humanizeValue(key)}: ${String(item)}`)
      .join(', ')
  } catch {
    return value
  }
}

function humanizeValue(value: string) {
  return value.replaceAll('_', ' ').replace(/\b\w/g, (letter) => letter.toUpperCase())
}

function enumValue<T extends string>(value: NullableValue<T> | T | null | undefined, fallback: T, keys: string[]) {
  if (!value) {
    return fallback
  }
  if (typeof value === 'string') {
    return value
  }
  const valid = value.Valid ?? value.valid ?? true
  if (!valid) {
    return fallback
  }

  for (const key of keys) {
    const nextValue = value[key]
    if (typeof nextValue === 'string' && nextValue !== '') {
      return nextValue as T
    }
  }

  return fallback
}

function leadStatus(lead: Lead): LeadStatus {
  return enumValue(lead.status, 'Follow_Up', ['lead_status', 'LeadStatus', 'status'])
}

function propertyStatus(property: Property): PropertyStatus {
  return enumChoice(property.status, ['available', 'sold', 'rented'], 'available')
}

function propertyType(property: Property): string {
  return enumValue(property.type, 'شقة', ['property_type', 'PropertyType', 'type'])
}

function callOutcome(call: Call): CallOutcome {
  return enumChoice(call.outcome, ['follow_up', 'qualified', 'closed', 'unqualified', 'no_answer'], 'follow_up')
}

function callSentiment(call: Call): string {
  return humanizeValue(enumChoice(call.sentiment, ['positive', 'negative', 'neutral'], 'neutral'))
}

function callOutcomeClass(outcome: CallOutcome) {
  return outcome === 'follow_up' ? 'follow_up' : outcome
}

function enumChoice<T extends string>(value: unknown, allowed: readonly T[], fallback: T): T {
  if (typeof value === 'string') {
    return normalizeChoice(value, allowed, fallback)
  }
  if (!value || typeof value !== 'object') {
    return fallback
  }

  const record = value as Record<string, unknown>
  const valid = record.Valid ?? record.valid ?? true
  if (valid === false) {
    return fallback
  }

  for (const item of Object.values(record)) {
    if (typeof item === 'string') {
      const choice = normalizeChoice(item, allowed, fallback)
      if (choice !== fallback || item.toLowerCase() === fallback.toLowerCase()) {
        return choice
      }
    }
  }

  return fallback
}

function normalizeChoice<T extends string>(value: string, allowed: readonly T[], fallback: T): T {
  const normalized = value.trim().toLowerCase()
  return allowed.find((item) => item.toLowerCase() === normalized) ?? fallback
}

function leadStatusClass(status: LeadStatus) {
  return status === 'Follow_Up' ? 'follow_up' : status
}

function valueLabel(value: unknown) {
  if (value === null || value === undefined || value === '') {
    return 'Not set'
  }
  return String(value)
}

function priceLabel(value: unknown) {
  if (value === null || value === undefined || value === '') {
    return 'No price'
  }
  return valueLabel(value)
}

function areaLabel(value: unknown) {
  if (value === null || value === undefined || value === '') {
    return 'No area'
  }
  return `${valueLabel(value)} sqm`
}

function numberValue(value: NullableNumber | number | null | undefined) {
  if (typeof value === 'number') {
    return value
  }
  if (!value) {
    return 0
  }
  const valid = value.Valid ?? value.valid ?? true
  if (!valid) {
    return 0
  }
  return value.Int32 ?? value.Int64 ?? value.int32 ?? value.int64 ?? 0
}

function durationLabel(value: NullableNumber | number | null | undefined) {
  const seconds = numberValue(value)
  if (seconds <= 0) {
    return 'No duration'
  }
  if (seconds < 60) {
    return `${seconds}s`
  }
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  return remainingSeconds === 0 ? `${minutes}m` : `${minutes}m ${remainingSeconds}s`
}

function dateLabel(value?: string) {
  if (!value) {
    return 'No date'
  }
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
}

export default App
