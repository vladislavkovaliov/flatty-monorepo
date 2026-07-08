import 'zone.js/dist/zone-node';
import { ngExpressEngine } from '@angular/ssr';
import express from 'express';
import { fileURLToPath } from 'node:url';
import { dirname, join, resolve } from 'node:path';
import bootstrap from './main.server';

const serverDistFolder = dirname(fileURLToPath(import.meta.url));
const browserDistFolder = resolve(serverDistFolder, '../browser');

const app = express();

app.engine('html', ngExpressEngine({ bootstrap }));
app.set('view engine', 'html');
app.set('views', browserDistFolder);

app.get(
  '*.*',
  express.static(browserDistFolder, { maxAge: '1y' })
);

app.get('*', (req, res) => {
  res.render('index', { req, providers: [] });
});

const port = process.env['PORT'] || 4000;
app.listen(port, () => {
  console.log(`Angular SSR listening on http://localhost:${port}`);
});
