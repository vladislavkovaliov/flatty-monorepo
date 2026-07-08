import './index.css';

import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { RouterProvider } from 'react-router-dom';

import {router} from './routers'

export const BootstrapApp = () => {
    return (
        <StrictMode>
            <RouterProvider router={router} />
        </StrictMode>
    )
};

createRoot(document.getElementById('root')!).render(<BootstrapApp />);
