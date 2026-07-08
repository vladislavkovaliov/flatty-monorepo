// NODE
import path from "node:path";
import { fileURLToPath } from "node:url";
//
import TsconfigPathsPlugin from 'tsconfig-paths-webpack-plugin';

const contextValue = './src';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const srcPathRelative = './src';
const srcPathAbsolute = path.join(__dirname, srcPathRelative);
const distPathRelative = 'dist';
const distPathAbsolute = path.resolve(__dirname, distPathRelative);

export default (env, argv) => {
    const {moveToPath} = env;
    const {mode} = argv;

    const isDevelopment = mode === 'development';

    let publicPathPrefix = "";

    if (!isDevelopment && moveToPath) {
        publicPathPrefix = `/${moveToPath}`;
    }

    const config = {
        mode: isDevelopment ? 'development' : 'production',
        context: path.resolve(__dirname, contextValue),
        resolve: {
            extensions: ['.ts', '.tsx', '.js', '.scss', '.css'],
            plugins: [
                new TsconfigPathsPlugin({
                    configFile: './tsconfig.webpack.config.json',
                }),
            ],
        },
        entry: {
            settings: `./main.tsx`,
        },
        output: {
            filename: '[name].js',
            chunkFilename: '[name].js',
            path: distPathAbsolute,
            publicPath: 'auto',
            chunkLoadingGlobal: 'jsonp',
            libraryTarget: 'umd',
            library: ['apps', '[name]'],
        },
        devtool: 'inline-source-map',
        module: {
            rules: [
                {
                    test: /\.(ts|tsx)$/,
                    use: {
                        loader: 'ts-loader',
                        options: {
                            configFile: path.resolve(__dirname, './tsconfig.webpack.config.json'),
                        },
                    },
                    exclude: /node_modules/,
                },
                {
                    enforce: 'pre',
                    test: /\.js$/,
                    loader: 'source-map-loader',
                },
                {
                    test: /\.css$/,
                    use: ['style-loader', 'css-loader'],
                },
                {
                    test: /\.(png|jpg|gif|svg)$/,
                    loader: 'file-loader',
                    options: {
                        outputPath: '',
                        ...(publicPathPrefix ? { publicPath: publicPathPrefix } : {}),
                        name: '[path][name].[ext]',
                    },
                },
            ],
        },
        devServer: {
            port: 8081,
        },
        // externals: {
        //     'react': {
        //         commonjs: 'react',
        //         commonjs2: 'react',
        //         amd: 'react',
        //         root: 'React',
        //     },
        //     'react-dom': {
        //         commonjs: 'react-dom',
        //         commonjs2: 'react-dom',
        //         amd: 'react-dom',
        //         root: 'ReactDOM',
        //     },
        //     // TODO: rest libs
        // },
        optimization: {
            splitChunks: {
                name: 'vendor',
                chunks: 'all',
                filename: 'vendor.js',
                minChunks: 2,
            }
        }
    };

    return config;
};

