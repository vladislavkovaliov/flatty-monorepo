import { useMutation } from '@tanstack/react-query';

type GqlAcceptInvitationData = { acceptInvitation: boolean };

async function graphqlRequest<T>(query: string, variables: Record<string, unknown>): Promise<T> {
  const response = await fetch('/graphql', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ query, variables }),
  });

  const json = await response.json();

  if (json.errors) {
    throw new Error(json.errors[0]?.message ?? 'GraphQL error');
  }

  return json.data as T;
}

const ACCEPT_INVITATION = `
  mutation AcceptInvitation($email: String!) {
    acceptInvitation(email: $email)
  }
`;

export function useAcceptInvitation() {
  return useMutation({
    mutationFn: (email: string) =>
      graphqlRequest<GqlAcceptInvitationData>(ACCEPT_INVITATION, { email }),
  });
}
