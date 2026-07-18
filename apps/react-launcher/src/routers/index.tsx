import {createBrowserRouter} from "react-router-dom";
import {App} from "../App.tsx";
import {Launcher} from "../Launcher.tsx";

export const router = createBrowserRouter([
    {
        path: '/',
        element: <App />,
        children: [
            {
                index: true,
                element: <Launcher />
            },
        ],
    },

]);