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
import { signUp } from '#/lib/auth-client'

export function RegisterPage() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setLoading(true)

    try {
      const { error: signUpError } = await signUp.email({ email, password, name })

      if (signUpError) {
        setError(signUpError.message ?? 'Registration failed')
        setLoading(false)
        return
      }

      window.location.href = '/'
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed')
      setLoading(false)
    }
  }

  return (
    <Center h="100vh">
      <Paper withBorder shadow="md" p="xl" radius="md" miw={400}>
        <form onSubmit={handleSubmit}>
          <Stack gap="md">
            <Title order={2}>Create Account</Title>

            {error && (
              <Alert color="red" variant="light">
                {error}
              </Alert>
            )}

            <TextInput
              label="Name"
              placeholder="Your name"
              value={name}
              onChange={(e) => setName(e.currentTarget.value)}
              required
            />

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
              Register
            </Button>

            <Text size="sm" ta="center">
              Already have an account? <Link to="/login">Sign In</Link>
            </Text>
          </Stack>
        </form>
      </Paper>
    </Center>
  )
}
