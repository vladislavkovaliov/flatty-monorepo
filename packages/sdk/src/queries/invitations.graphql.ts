import { useMutation } from '@tanstack/react-query';
import { graphqlRequest } from '../lib/graphql';

type GqlAcceptInvitationData = { acceptInvitation: boolean };

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
