//
// NOTE: This is an integration-level test requiring a running server.
// It uses supertest + the NestJS testing utilities.
// Install supertest: `npm install -D supertest @types/supertest`
//
// import { describe, it, expect, beforeAll } from 'vitest'
// import supertest from 'supertest'
// import { bootstrap } from '../main'
//
// describe('POST /api/auth/set-password', () => {
//   let request: ReturnType<typeof supertest>
//
//   beforeAll(async () => {
//     const app = await bootstrap()
//     request = supertest(app.getHttpServer())
//   })
//
//   it('returns 400 when no password is provided', async () => {
//     const res = await request
//       .post('/api/auth/set-password')
//       .send({})
//       .expect(400)
//     expect(res.body.error).toBe('Password is required')
//   })
//
//   it('returns 401 when not authenticated', async () => {
//     const res = await request
//       .post('/api/auth/set-password')
//       .send({ newPassword: 'somePassword123' })
//       .expect(401)
//     expect(res.body.error).toBe('Not authenticated')
//   })
// })
