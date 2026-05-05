import { useCallback, useEffect, useMemo, useState } from 'react'
import type { FormEvent, SVGProps } from 'react'
import './App.css'

type ApiResponse<T> = {
  is_ok: boolean
  message: string
  payload: T
}

type Tokens = {
  access: string
  refresh: string
}

type Lead = {
  id: string
  name: string
  title: string
  description: string
  contact: string
  status: string
  created_at: string
  updated_at: string
}

type LeadsList = {
  items: Lead[]
  count: number
  limit: number
  offset: number
}

type LeadForm = {
  name: string
  title: string
  contact: string
  description: string
}

type LoginForm = {
  login: string
  password: string
}

type Filters = {
  q: string
  status: string
  date_from: string
  date_to: string
  sort: string
  order: 'ASC' | 'DESC'
  limit: number
  offset: number
}

type Page = 'lead' | 'admin'
type NoticeKind = 'idle' | 'success' | 'error' | 'info'

type Notice = {
  kind: NoticeKind
  text: string
}

type IconName =
  | 'arrowLeft'
  | 'arrowRight'
  | 'briefcase'
  | 'check'
  | 'filter'
  | 'inbox'
  | 'lock'
  | 'logout'
  | 'refresh'
  | 'send'
  | 'spark'
  | 'user'

const tokenStorageKey = 'leadflow-admin-tokens'

const initialLeadForm: LeadForm = {
  name: '',
  title: '',
  contact: '',
  description: '',
}

const initialLoginForm: LoginForm = {
  login: '',
  password: '',
}

const initialFilters: Filters = {
  q: '',
  status: '',
  date_from: '',
  date_to: '',
  sort: 'created_at',
  order: 'DESC',
  limit: 20,
  offset: 0,
}

const statusOptions = [
  { value: '', label: 'Все статусы' },
  { value: 'new', label: 'Новые' },
  { value: 'in_progress', label: 'В работе' },
  { value: 'done', label: 'Закрытые' },
  { value: 'rejected', label: 'Отклоненные' },
]

const statusActions = [
  { value: 'new', label: 'Новый' },
  { value: 'in_progress', label: 'В работе' },
  { value: 'done', label: 'Закрыт' },
  { value: 'rejected', label: 'Отклонен' },
]

const sortOptions = [
  { value: 'created_at', label: 'По дате создания' },
  { value: 'updated_at', label: 'По обновлению' },
  { value: 'name', label: 'По имени' },
  { value: 'status', label: 'По статусу' },
]

class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.status = status
  }
}

function App() {
  const [page, setPage] = useState<Page>(() => getPageFromLocation())
  const [leadForm, setLeadForm] = useState<LeadForm>(initialLeadForm)
  const [leadNotice, setLeadNotice] = useState<Notice>({ kind: 'idle', text: '' })
  const [isLeadSubmitting, setIsLeadSubmitting] = useState(false)

  const [tokens, setTokens] = useState<Tokens | null>(() => readStoredTokens())
  const [loginForm, setLoginForm] = useState<LoginForm>(initialLoginForm)
  const [adminNotice, setAdminNotice] = useState<Notice>({ kind: 'idle', text: '' })
  const [isLoginSubmitting, setIsLoginSubmitting] = useState(false)
  const [isLeadsLoading, setIsLeadsLoading] = useState(false)
  const [isLogoutLoading, setIsLogoutLoading] = useState(false)
  const [statusInFlight, setStatusInFlight] = useState<string | null>(null)
  const [filters, setFilters] = useState<Filters>(initialFilters)
  const [appliedFilters, setAppliedFilters] = useState<Filters>(initialFilters)
  const [leads, setLeads] = useState<LeadsList>({
    items: [],
    count: 0,
    limit: initialFilters.limit,
    offset: initialFilters.offset,
  })

  const hasActiveFilters = useMemo(
    () => Boolean(filters.q || filters.status || filters.date_from || filters.date_to),
    [filters],
  )

  useEffect(() => {
    const onPopState = () => setPage(getPageFromLocation())
    window.addEventListener('popstate', onPopState)
    return () => window.removeEventListener('popstate', onPopState)
  }, [])

  const saveTokens = useCallback((nextTokens: Tokens | null) => {
    setTokens(nextTokens)
    if (nextTokens) {
      localStorage.setItem(tokenStorageKey, JSON.stringify(nextTokens))
      return
    }
    localStorage.removeItem(tokenStorageKey)
  }, [])

  const requestWithAuth = useCallback(async <T,>(
    currentTokens: Tokens,
    path: string,
    options: RequestInit = {},
  ): Promise<T> => {
    try {
      return await apiRequest<T>(path, options, currentTokens.access)
    } catch (error) {
      if (!(error instanceof ApiError) || error.status !== 401) {
        throw error
      }

      const refreshed = await apiRequest<Tokens>('/api/admin/refresh', {
        method: 'POST',
        body: JSON.stringify({ refresh: currentTokens.refresh }),
      })
      saveTokens(refreshed)
      return apiRequest<T>(path, options, refreshed.access)
    }
  }, [saveTokens])

  const loadLeads = useCallback(async (currentTokens: Tokens, currentFilters: Filters) => {
    setIsLeadsLoading(true)
    setAdminNotice({ kind: 'info', text: 'Загружаем заявки' })

    try {
      const params = new URLSearchParams()
      Object.entries(currentFilters).forEach(([key, value]) => {
        if (value !== '' && value !== null && value !== undefined) {
          params.set(key, String(value))
        }
      })

      const payload = await requestWithAuth<LeadsList>(
        currentTokens,
        `/api/admin/leads?${params.toString()}`,
      )
      setLeads(payload)
      setAdminNotice({
        kind: payload.items.length ? 'success' : 'info',
        text: payload.items.length ? 'Заявки обновлены' : 'По текущим фильтрам заявок нет',
      })
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        saveTokens(null)
      }
      setAdminNotice({ kind: 'error', text: messageFromError(error) })
    } finally {
      setIsLeadsLoading(false)
    }
  }, [requestWithAuth, saveTokens])

  useEffect(() => {
    let isCurrent = true
    if (page === 'admin' && tokens) {
      queueMicrotask(() => {
        if (isCurrent) {
          void loadLeads(tokens, appliedFilters)
        }
      })
    }

    return () => {
      isCurrent = false
    }
  }, [appliedFilters, loadLeads, page, tokens])

  const handleLeadSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setIsLeadSubmitting(true)
    setLeadNotice({ kind: 'info', text: 'Отправляем заявку' })

    try {
      const payload = await apiRequest<{ id: string }>('/api/lead', {
        method: 'POST',
        body: JSON.stringify(trimLeadForm(leadForm)),
      })
      setLeadForm(initialLeadForm)
      setLeadNotice({
        kind: 'success',
        text: `Заявка принята. Номер: ${payload.id.slice(0, 8)}`,
      })
    } catch (error) {
      setLeadNotice({ kind: 'error', text: messageFromError(error) })
    } finally {
      setIsLeadSubmitting(false)
    }
  }

  const handleLoginSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setIsLoginSubmitting(true)
    setAdminNotice({ kind: 'info', text: 'Проверяем доступ' })

    try {
      const payload = await apiRequest<Tokens>('/api/admin/login', {
        method: 'POST',
        body: JSON.stringify({
          login: loginForm.login.trim(),
          password: loginForm.password.trim(),
        }),
      })
      setLoginForm(initialLoginForm)
      saveTokens(payload)
      setAdminNotice({ kind: 'success', text: 'Доступ открыт' })
    } catch (error) {
      setAdminNotice({ kind: 'error', text: messageFromError(error) })
    } finally {
      setIsLoginSubmitting(false)
    }
  }

  const handleLogout = async () => {
    if (!tokens) {
      return
    }

    setIsLogoutLoading(true)
    try {
      await requestWithAuth<null>(tokens, '/api/admin/logout', {
        method: 'POST',
        body: JSON.stringify({ refresh: tokens.refresh }),
      })
    } catch {
      // Local logout still clears an unusable or already-revoked session.
    } finally {
      saveTokens(null)
      setLeads({ items: [], count: 0, limit: initialFilters.limit, offset: 0 })
      setAdminNotice({ kind: 'idle', text: '' })
      setIsLogoutLoading(false)
    }
  }

  const handleFiltersSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!tokens) {
      return
    }

    const nextFilters = { ...filters, offset: 0 }
    setFilters(nextFilters)
    setAppliedFilters(nextFilters)
  }

  const handleResetFilters = () => {
    setFilters(initialFilters)
    setAppliedFilters(initialFilters)
  }

  const handlePageChange = (direction: 'prev' | 'next') => {
    const nextOffset =
      direction === 'next'
        ? appliedFilters.offset + appliedFilters.limit
        : Math.max(0, appliedFilters.offset - appliedFilters.limit)
    const nextFilters = { ...appliedFilters, offset: nextOffset }
    setFilters(nextFilters)
    setAppliedFilters(nextFilters)
  }

  const handleStatusChange = async (lead: Lead, status: string) => {
    if (!tokens || status === lead.status) {
      return
    }

    setStatusInFlight(lead.id)
    setAdminNotice({ kind: 'info', text: 'Обновляем статус' })

    try {
      await requestWithAuth<null>(tokens, `/api/admin/leads/${lead.id}/status`, {
        method: 'PATCH',
        body: JSON.stringify({ status }),
      })
      setLeads((current) => ({
        ...current,
        items: current.items.map((item) =>
          item.id === lead.id
            ? { ...item, status, updated_at: new Date().toISOString() }
            : item,
        ),
      }))
      setAdminNotice({ kind: 'success', text: 'Статус обновлен' })
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        saveTokens(null)
      }
      setAdminNotice({ kind: 'error', text: messageFromError(error) })
    } finally {
      setStatusInFlight(null)
    }
  }

  return (
    <div className="app-shell">
      <main className="workspace">
        {page === 'lead' ? (
          <LeadPage
            form={leadForm}
            notice={leadNotice}
            isSubmitting={isLeadSubmitting}
            onChange={setLeadForm}
            onSubmit={handleLeadSubmit}
          />
        ) : (
          <AdminPage
            adminNotice={adminNotice}
            filters={filters}
            hasActiveFilters={hasActiveFilters}
            isLeadsLoading={isLeadsLoading}
            isLoginSubmitting={isLoginSubmitting}
            isLogoutLoading={isLogoutLoading}
            leads={leads}
            loginForm={loginForm}
            statusInFlight={statusInFlight}
            tokens={tokens}
            onFiltersChange={setFilters}
            onFiltersReset={handleResetFilters}
            onFiltersSubmit={handleFiltersSubmit}
            onLoginChange={setLoginForm}
            onLoginSubmit={handleLoginSubmit}
            onLogout={handleLogout}
            onPageChange={handlePageChange}
            onStatusChange={handleStatusChange}
          />
        )}
      </main>
    </div>
  )
}

function LeadPage({
  form,
  notice,
  isSubmitting,
  onChange,
  onSubmit,
}: {
  form: LeadForm
  notice: Notice
  isSubmitting: boolean
  onChange: (form: LeadForm) => void
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
}) {
  return (
    <section className="lead-layout">
      <div className="lead-copy">
        <span className="section-kicker">Публичная форма</span>
        <h1>Оставьте заявку, мы быстро передадим ее команде</h1>
        <p>
          Соберите имя, контакт, тему и подробности задачи в одном коротком
          сценарии. После отправки lead сразу появится в админском списке.
        </p>

        <div className="process-strip" aria-label="Этапы обработки">
          <div>
            <span>01</span>
            <strong>Заявка</strong>
          </div>
          <div>
            <span>02</span>
            <strong>Контакт</strong>
          </div>
          <div>
            <span>03</span>
            <strong>Статус</strong>
          </div>
        </div>
      </div>

      <form className="lead-form panel" onSubmit={onSubmit}>
        <div className="panel-heading">
          <div>
            <span className="section-kicker">Lead intake</span>
            <h2>Новая заявка</h2>
          </div>
          <span className="panel-icon" aria-hidden="true">
            <Icon name="send" />
          </span>
        </div>

        <div className="field-grid">
          <label className="field">
            <span>Имя</span>
            <input
              autoComplete="name"
              name="name"
              placeholder="Анна Смирнова"
              required
              value={form.name}
              onChange={(event) => onChange({ ...form, name: event.target.value })}
            />
          </label>

          <label className="field">
            <span>Контакт</span>
            <input
              autoComplete="email"
              name="contact"
              placeholder="anna@example.com"
              required
              value={form.contact}
              onChange={(event) => onChange({ ...form, contact: event.target.value })}
            />
          </label>
        </div>

        <label className="field">
          <span>Тема</span>
          <input
            name="title"
            placeholder="Запуск лендинга"
            required
            value={form.title}
            onChange={(event) => onChange({ ...form, title: event.target.value })}
          />
        </label>

        <label className="field">
          <span>Описание</span>
          <textarea
            name="description"
            placeholder="Коротко опишите задачу, сроки и важные детали"
            required
            rows={6}
            value={form.description}
            onChange={(event) =>
              onChange({ ...form, description: event.target.value })
            }
          />
        </label>

        <div className="form-footer">
          <NoticeMessage notice={notice} />
          <button className="primary-button" disabled={isSubmitting} type="submit">
            <Icon name={isSubmitting ? 'refresh' : 'send'} />
            <span>{isSubmitting ? 'Отправляем' : 'Отправить'}</span>
          </button>
        </div>
      </form>
    </section>
  )
}

function AdminPage({
  adminNotice,
  filters,
  hasActiveFilters,
  isLeadsLoading,
  isLoginSubmitting,
  isLogoutLoading,
  leads,
  loginForm,
  statusInFlight,
  tokens,
  onFiltersChange,
  onFiltersReset,
  onFiltersSubmit,
  onLoginChange,
  onLoginSubmit,
  onLogout,
  onPageChange,
  onStatusChange,
}: {
  adminNotice: Notice
  filters: Filters
  hasActiveFilters: boolean
  isLeadsLoading: boolean
  isLoginSubmitting: boolean
  isLogoutLoading: boolean
  leads: LeadsList
  loginForm: LoginForm
  statusInFlight: string | null
  tokens: Tokens | null
  onFiltersChange: (filters: Filters) => void
  onFiltersReset: () => void
  onFiltersSubmit: (event: FormEvent<HTMLFormElement>) => void
  onLoginChange: (form: LoginForm) => void
  onLoginSubmit: (event: FormEvent<HTMLFormElement>) => void
  onLogout: () => void
  onPageChange: (direction: 'prev' | 'next') => void
  onStatusChange: (lead: Lead, status: string) => void
}) {
  if (!tokens) {
    return (
      <section className="admin-login-layout">
        <div className="login-copy">
          <span className="section-kicker">Admin access</span>
          <h1>Единая страница для входа и управления lead’ами</h1>
          <p>
            После успешного входа форма логина на этом же маршруте заменяется
            рабочей зоной со списком заявок, фильтрами и сменой статусов.
          </p>
        </div>

        <form className="login-card panel" onSubmit={onLoginSubmit}>
          <div className="panel-heading">
            <div>
              <span className="section-kicker">Защищенная зона</span>
              <h2>Вход администратора</h2>
            </div>
            <span className="panel-icon" aria-hidden="true">
              <Icon name="lock" />
            </span>
          </div>

          <label className="field">
            <span>Логин</span>
            <input
              autoComplete="username"
              name="login"
              placeholder="admin"
              required
              value={loginForm.login}
              onChange={(event) =>
                onLoginChange({ ...loginForm, login: event.target.value })
              }
            />
          </label>

          <label className="field">
            <span>Пароль</span>
            <input
              autoComplete="current-password"
              name="password"
              placeholder="Пароль"
              required
              type="password"
              value={loginForm.password}
              onChange={(event) =>
                onLoginChange({ ...loginForm, password: event.target.value })
              }
            />
          </label>

          <div className="form-footer">
            <NoticeMessage notice={adminNotice} />
            <button className="primary-button" disabled={isLoginSubmitting} type="submit">
              <Icon name={isLoginSubmitting ? 'refresh' : 'user'} />
              <span>{isLoginSubmitting ? 'Входим' : 'Войти'}</span>
            </button>
          </div>
        </form>
      </section>
    )
  }

  return (
    <section className="admin-dashboard">
      <header className="dashboard-header">
        <div>
          <span className="section-kicker">Admin dashboard</span>
          <h1>Управление lead’ами</h1>
        </div>
        <button
          className="secondary-button"
          disabled={isLogoutLoading}
          type="button"
          onClick={onLogout}
        >
          <Icon name="logout" />
          <span>{isLogoutLoading ? 'Выходим' : 'Выйти'}</span>
        </button>
      </header>

      <form className="filters-bar panel" onSubmit={onFiltersSubmit}>
        <label className="field compact-field search-field">
          <span>Поиск</span>
          <input
            placeholder="Имя, тема, контакт"
            value={filters.q}
            onChange={(event) =>
              onFiltersChange({ ...filters, q: event.target.value, offset: 0 })
            }
          />
        </label>

        <label className="field compact-field">
          <span>Статус</span>
          <select
            value={filters.status}
            onChange={(event) =>
              onFiltersChange({ ...filters, status: event.target.value, offset: 0 })
            }
          >
            {statusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>

        <label className="field compact-field">
          <span>С</span>
          <input
            type="date"
            value={filters.date_from}
            onChange={(event) =>
              onFiltersChange({ ...filters, date_from: event.target.value, offset: 0 })
            }
          />
        </label>

        <label className="field compact-field">
          <span>По</span>
          <input
            type="date"
            value={filters.date_to}
            onChange={(event) =>
              onFiltersChange({ ...filters, date_to: event.target.value, offset: 0 })
            }
          />
        </label>

        <label className="field compact-field">
          <span>Сортировка</span>
          <select
            value={filters.sort}
            onChange={(event) => onFiltersChange({ ...filters, sort: event.target.value })}
          >
            {sortOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>

        <label className="field compact-field">
          <span>Порядок</span>
          <select
            value={filters.order}
            onChange={(event) =>
              onFiltersChange({ ...filters, order: event.target.value as Filters['order'] })
            }
          >
            <option value="DESC">Сначала новые</option>
            <option value="ASC">Сначала старые</option>
          </select>
        </label>

        <div className="filters-actions">
          <button className="icon-button" title="Применить фильтры" type="submit">
            <Icon name="filter" />
          </button>
          <button
            className="icon-button muted"
            disabled={!hasActiveFilters}
            title="Сбросить фильтры"
            type="button"
            onClick={onFiltersReset}
          >
            <Icon name="refresh" />
          </button>
        </div>
      </form>

      <div className="dashboard-meta">
        <NoticeMessage notice={adminNotice} />
        <div className="pager">
          <button
            className="icon-button"
            disabled={filters.offset === 0 || isLeadsLoading}
            title="Предыдущая страница"
            type="button"
            onClick={() => onPageChange('prev')}
          >
            <Icon name="arrowLeft" />
          </button>
          <span>
            {filters.offset + 1}-{filters.offset + leads.count}
          </span>
          <button
            className="icon-button"
            disabled={leads.count < filters.limit || isLeadsLoading}
            title="Следующая страница"
            type="button"
            onClick={() => onPageChange('next')}
          >
            <Icon name="arrowRight" />
          </button>
        </div>
      </div>

      {leads.items.length ? (
        <div className={isLeadsLoading ? 'leads-grid dimmed' : 'leads-grid'}>
          {leads.items.map((lead) => (
            <article className="lead-card" key={lead.id}>
              <header>
                <div>
                  <span className="lead-id">#{lead.id.slice(0, 8)}</span>
                  <h2>{lead.title}</h2>
                </div>
                <span className={`status-pill ${statusClass(lead.status)}`}>
                  {statusLabel(lead.status)}
                </span>
              </header>

              <p>{lead.description}</p>

              <dl className="lead-details">
                <div>
                  <dt>Клиент</dt>
                  <dd>{lead.name}</dd>
                </div>
                <div>
                  <dt>Контакт</dt>
                  <dd>{lead.contact}</dd>
                </div>
                <div>
                  <dt>Создано</dt>
                  <dd>{formatDate(lead.created_at)}</dd>
                </div>
                <div>
                  <dt>Обновлено</dt>
                  <dd>{formatDate(lead.updated_at)}</dd>
                </div>
              </dl>

              <label className="field compact-field">
                <span>Сменить статус</span>
                <select
                  disabled={statusInFlight === lead.id}
                  value={lead.status}
                  onChange={(event) => onStatusChange(lead, event.target.value)}
                >
                  {statusActions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                  {isKnownStatus(lead.status) ? null : (
                    <option value={lead.status}>{lead.status}</option>
                  )}
                </select>
              </label>
            </article>
          ))}
        </div>
      ) : (
        <div className="empty-state panel">
          <Icon name="inbox" />
          <h2>Заявок пока нет</h2>
          <p>Когда пользователи отправят форму, новые lead’ы появятся здесь.</p>
        </div>
      )}
    </section>
  )
}

function NoticeMessage({ notice }: { notice: Notice }) {
  if (!notice.text || notice.kind === 'idle') {
    return <span className="notice-placeholder" aria-hidden="true" />
  }

  return (
    <p className={`notice ${notice.kind}`} role={notice.kind === 'error' ? 'alert' : 'status'}>
      {notice.kind === 'success' ? <Icon name="check" /> : null}
      {notice.kind === 'error' ? <Icon name="lock" /> : null}
      {notice.kind === 'info' ? <Icon name="refresh" /> : null}
      <span>{notice.text}</span>
    </p>
  )
}

function Icon({ name, ...props }: { name: IconName } & SVGProps<SVGSVGElement>) {
  return (
    <svg
      aria-hidden="true"
      fill="none"
      focusable="false"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="2"
      viewBox="0 0 24 24"
      {...props}
    >
      {iconPath(name)}
    </svg>
  )
}

function iconPath(name: IconName) {
  switch (name) {
    case 'arrowLeft':
      return <path d="M19 12H5m6-6-6 6 6 6" />
    case 'arrowRight':
      return <path d="M5 12h14m-6-6 6 6-6 6" />
    case 'briefcase':
      return <path d="M10 6V5a2 2 0 0 1 2-2h0a2 2 0 0 1 2 2v1m5 0H5a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2Zm-7 7h.01" />
    case 'check':
      return <path d="m20 6-11 11-5-5" />
    case 'filter':
      return <path d="M4 5h16M7 12h10m-7 7h4" />
    case 'inbox':
      return <path d="M22 12h-6l-2 3h-4l-2-3H2m20 0-3-7H5l-3 7v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2Z" />
    case 'lock':
      return <path d="M7 11V8a5 5 0 0 1 10 0v3M6 11h12v9H6Z" />
    case 'logout':
      return <path d="M10 17 5 12l5-5m-5 5h12M14 4h4a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2h-4" />
    case 'refresh':
      return <path d="M20 11a8 8 0 0 0-14-5L4 8m0 0h5M4 8V3m0 10a8 8 0 0 0 14 5l2-2m0 0h-5m5 0v5" />
    case 'send':
      return <path d="m22 2-7 20-4-9-9-4Z M22 2 11 13" />
    case 'spark':
      return <path d="M13 2 9 11l-7 2 7 2 4 7 3-7 6-2-6-2Z" />
    case 'user':
      return <path d="M20 21a8 8 0 0 0-16 0m12-13a4 4 0 1 1-8 0 4 4 0 0 1 8 0Z" />
  }
}

async function apiRequest<T>(path: string, options: RequestInit = {}, accessToken?: string) {
  const headers = new Headers(options.headers)
  headers.set('Content-Type', 'application/json')
  if (accessToken) {
    headers.set('Authorization', `Bearer ${accessToken}`)
  }

  const response = await fetch(path, {
    ...options,
    headers,
  })
  const data = (await response.json().catch(() => null)) as ApiResponse<T> | null

  if (!response.ok || !data?.is_ok) {
    throw new ApiError(data?.message || 'Не удалось выполнить запрос', response.status)
  }

  return data.payload
}

function readStoredTokens() {
  const stored = localStorage.getItem(tokenStorageKey)
  if (!stored) {
    return null
  }

  try {
    const parsed = JSON.parse(stored) as Partial<Tokens>
    if (parsed.access && parsed.refresh) {
      return { access: parsed.access, refresh: parsed.refresh }
    }
  } catch {
    localStorage.removeItem(tokenStorageKey)
  }

  return null
}

function getPageFromLocation(): Page {
  return window.location.pathname.startsWith('/admin') ? 'admin' : 'lead'
}

function trimLeadForm(form: LeadForm) {
  return {
    name: form.name.trim(),
    title: form.title.trim(),
    contact: form.contact.trim(),
    description: form.description.trim(),
  }
}

function messageFromError(error: unknown) {
  if (error instanceof ApiError) {
    return apiMessage(error.message)
  }
  if (error instanceof TypeError) {
    return 'Не удалось подключиться к API'
  }
  if (error instanceof Error) {
    return error.message
  }
  return 'Что-то пошло не так'
}

function apiMessage(message: string) {
  const messages: Record<string, string> = {
    'database error': 'Сервис временно не смог сохранить данные',
    'database login error': 'Логин или пароль не подошли',
    'database session error': 'Сессия устарела, войдите заново',
    'invalid query': 'Проверьте фильтры и даты',
    'invalid request': 'Проверьте заполненные поля',
    'invalid values': 'Заполните обязательные поля корректно',
    'jwt error': 'Не удалось создать сессию',
    'not found': 'Заявка не найдена',
    'token expired': 'Сессия истекла, войдите заново',
    unauthorized: 'Нет доступа, войдите заново',
  }

  return messages[message] || message
}

function statusLabel(status: string) {
  const labels: Record<string, string> = {
    done: 'Закрыт',
    in_progress: 'В работе',
    new: 'Новый',
    rejected: 'Отклонен',
  }

  return labels[status] || status
}

function statusClass(status: string) {
  if (status === 'done') {
    return 'done'
  }
  if (status === 'in_progress') {
    return 'progress'
  }
  if (status === 'rejected') {
    return 'rejected'
  }
  return 'new'
}

function isKnownStatus(status: string) {
  return statusActions.some((option) => option.value === status)
}

function formatDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }

  return new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    month: 'short',
  }).format(date)
}

export default App
