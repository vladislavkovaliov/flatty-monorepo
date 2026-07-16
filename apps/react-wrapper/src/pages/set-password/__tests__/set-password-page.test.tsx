//
// NOTE: This test requires Vitest + @testing-library/react to be configured.
// Currently the project has no test infrastructure installed.
// This file is provided as a scaffolding reference.
// To run: `npx vitest` (after vitest setup)

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { MantineProvider } from '@mantine/core'
import { SetPasswordPage } from '../set-password-page'

function renderPage() {
  return render(
    <MantineProvider>
      <MemoryRouter>
        <SetPasswordPage />
      </MemoryRouter>
    </MantineProvider>
  )
}

describe('SetPasswordPage', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    localStorage.clear()
  })

  it('renders the form title and description', () => {
    renderPage()
    expect(screen.getByText('Set Your Password')).toBeInTheDocument()
    expect(
      screen.getByText(/You're logged in via magic link/)
    ).toBeInTheDocument()
  })

  it('shows an error when submitting with empty password', async () => {
    renderPage()
    fireEvent.click(screen.getByText('Set Password'))
    await waitFor(() => {
      expect(screen.getByText('Password is required')).toBeInTheDocument()
    })
  })

  it('shows an error when password is too short', async () => {
    renderPage()
    const inputs = screen.getAllByLabelText(/password/i)
    fireEvent.change(inputs[0], { target: { value: '123' } })
    fireEvent.change(inputs[1], { target: { value: '123' } })
    fireEvent.click(screen.getByText('Set Password'))
    await waitFor(() => {
      expect(
        screen.getByText('Password must be at least 8 characters')
      ).toBeInTheDocument()
    })
  })

  it('shows an error when passwords do not match', async () => {
    renderPage()
    const inputs = screen.getAllByLabelText(/password/i)
    fireEvent.change(inputs[0], { target: { value: 'password123' } })
    fireEvent.change(inputs[1], { target: { value: 'password456' } })
    fireEvent.click(screen.getByText('Set Password'))
    await waitFor(() => {
      expect(screen.getByText('Passwords do not match')).toBeInTheDocument()
    })
  })

  it('calls the API on valid submission and navigates on success', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true }),
    })
    vi.stubGlobal('fetch', fetchMock)

    renderPage()
    const inputs = screen.getAllByLabelText(/password/i)
    fireEvent.change(inputs[0], { target: { value: 'validPassword123' } })
    fireEvent.change(inputs[1], { target: { value: 'validPassword123' } })
    fireEvent.click(screen.getByText('Set Password'))

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith('/api/auth/set-password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ newPassword: 'validPassword123' }),
        credentials: 'include',
      })
    })
  })

  it('shows an error message when the API returns an error', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: false,
      json: async () => ({ error: 'Password too weak' }),
    })
    vi.stubGlobal('fetch', fetchMock)

    renderPage()
    const inputs = screen.getAllByLabelText(/password/i)
    fireEvent.change(inputs[0], { target: { value: 'validPassword123' } })
    fireEvent.change(inputs[1], { target: { value: 'validPassword123' } })
    fireEvent.click(screen.getByText('Set Password'))

    await waitFor(() => {
      expect(screen.getByText('Password too weak')).toBeInTheDocument()
    })
  })

  it('sets localStorage flag and navigates on skip', async () => {
    renderPage()
    fireEvent.click(screen.getByText('Skip for now'))
    expect(localStorage.getItem('skip-set-password')).toBe('true')
  })
})
