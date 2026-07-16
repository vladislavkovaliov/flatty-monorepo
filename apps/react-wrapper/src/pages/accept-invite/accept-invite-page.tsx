import { useEffect } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import {
  Center,
  Paper,
  Stack,
  Title,
  Text,
  Loader,
  Alert,
  Button,
} from '@mantine/core'
import { useAcceptInvitation } from '@flatty-budget/sdk'
import { useSession } from '#/lib/auth-client'

export function AcceptInvitePage() {
  const navigate = useNavigate()
  const { data: session, isPending: isSessionLoading } = useSession()
  const { mutate, isPending, isSuccess, isError, error, reset } = useAcceptInvitation()

  useEffect(() => {
    if (!isSessionLoading && session?.user?.email) {
      mutate(session.user.email)
    }
  }, [isSessionLoading, session?.user?.email, mutate])

  if (isSessionLoading) {
    return (
      <Center h="100vh">
        <Stack align="center" gap="md">
          <Loader size="lg" />
          <Text c="dimmed">Checking your session...</Text>
        </Stack>
      </Center>
    )
  }

  if (!session) {
    return (
      <Center h="100vh">
        <Paper withBorder shadow="md" p="xl" radius="md" miw={400}>
          <Stack align="center" gap="md">
            <Title order={2}>Sign In Required</Title>
            <Text ta="center" c="dimmed">
              You need to be signed in to accept an invitation.
            </Text>
            <Button component={Link} to="/login" fullWidth>
              Sign In
            </Button>
          </Stack>
        </Paper>
      </Center>
    )
  }

  return (
    <Center h="100vh">
      <Paper withBorder shadow="md" p="xl" radius="md" miw={400}>
        <Stack align="center" gap="md">
          {isPending && (
            <>
              <Loader size="lg" />
              <Text c="dimmed">Accepting invitation...</Text>
            </>
          )}

          {isSuccess && (
            <>
              <Title order={2} c="green">Invitation Accepted!</Title>
              <Text ta="center" c="dimmed">
                You have successfully joined the household.
              </Text>
              <Button onClick={() => navigate('/set-password')} fullWidth>
                Go to Dashboard
              </Button>
            </>
          )}

          {isError && (
            <>
              <Alert color="red" variant="light" title="Error">
                {error instanceof Error ? error.message : 'Failed to accept invitation'}
              </Alert>
              <Button onClick={() => reset()} fullWidth variant="outline">
                Try Again
              </Button>
            </>
          )}
        </Stack>
      </Paper>
    </Center>
  )
}
