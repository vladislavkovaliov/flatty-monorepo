import './index.css';

import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { RouterProvider } from 'react-router-dom';

import {router} from './routers'
import { QueryProvider } from './app/providers/query-provider';
import { ThemeProvider } from './app/providers/theme-provider';

export const BootstrapApp = () => {
    return (
        <StrictMode>
            <QueryProvider>
                <ThemeProvider>
                    <RouterProvider router={router} />
                </ThemeProvider>
            </QueryProvider>
        </StrictMode>
    )
};

createRoot(document.getElementById('root')!).render(<BootstrapApp />);
