import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import {
  Stack,
  TextInput,
  PasswordInput,
  Button,
  Title,
  Text,
  Paper,
  Center,
  Alert,
} from '@mantine/core'
import { signIn } from '#/lib/auth-client'

export function LoginPage() {
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setLoading(true)

    try {
      const { error: signInError } = await signIn.email({ email, password })

      if (signInError) {
        setError(signInError.message ?? 'Invalid credentials')
        setLoading(false)
        return
      }

      window.location.href = '/'
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Sign in failed')
      setLoading(false)
    }
  }

  return (
    <Center h="100vh">
      <Paper withBorder shadow="md" p="xl" radius="md" miw={400}>
        <form onSubmit={handleSubmit}>
          <Stack gap="md">
            <Title order={2}>Sign In</Title>

            {error && (
              <Alert color="red" variant="light">
                {error}
              </Alert>
            )}

            <TextInput
              label="Email"
              placeholder="your@email.com"
              value={email}
              onChange={(e) => setEmail(e.currentTarget.value)}
              required
            />

            <PasswordInput
              label="Password"
              placeholder="Your password"
              value={password}
              onChange={(e) => setPassword(e.currentTarget.value)}
              required
            />

            <Button type="submit" fullWidth loading={loading}>
              Sign In
            </Button>

            <Text size="sm" ta="center">
              Don&apos;t have an account? <Link to="/register">Register</Link>
            </Text>
          </Stack>
        </form>
      </Paper>
    </Center>
  )
}
