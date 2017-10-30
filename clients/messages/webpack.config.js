const webpack = require('webpack');
const ExtractTextPlugin = require('extract-text-webpack-plugin');
const HtmlWebpackPlugin = require('html-webpack-plugin');

const entries = {
    index: './src/entries/index.tsx',
    app: './src/entries/app.tsx'
};

var config = Object.keys(entries).map(function(entry) {
    return {
        entry: entries[entry],

        output: {
            path: __dirname + '/dist',
            filename: entry + '-bundle.min.js'
        },

        devtool: 'source-map',

        resolve: {
            extensions: ['.js', '.json', '.ts', '.tsx', '.scss', '.css'],
            modules: ['node_modules', 'src']
        },

        module: {
            rules: [
                {
                    test: /\.ts$|\.tsx$/,
                    exclude: /node_modules/,
                    use: [
                        {
                            loader: 'babel-loader',
                            options: {
                                presets: ['es2015', 'react'],
                                sourceMap: true
                            }
                        },
                        {
                            loader: 'ts-loader'
                        }
                    ]
                },
                {
                    test: /\.css$|\.scss$/,
                    exclude: /node_modules\/(?!bootstrap\/).*/,
                    use: ExtractTextPlugin.extract({
                        use: [
                            {
                                loader: 'css-loader',
                                options: {
                                    sourceMap: true,
                                    minimize: true
                                }
                            },
                            {
                                loader: 'sass-loader',
                                options: {
                                    includePaths: [
                                        __dirname + '/src/stylesheets/sass',
                                        __dirname + '/node_modules/compass-mixins/lib'
                                    ],
                                    sourceMap: true
                                }
                            }
                        ],
                        fallback: 'style-loader'
                    })
                }
            ]
        },

        plugins: [
            new webpack.optimize.UglifyJsPlugin({
                minimize: true,
                sourceMap: true
            }),

            new webpack.optimize.CommonsChunkPlugin({
                name: 'commons',
                filename: 'commons.min.js'
            }),

            new webpack.DefinePlugin({
                'process.env': {
                    NODE_ENV: JSON.stringify('production')
                }
            }),

            new HtmlWebpackPlugin({
                title: 'Webpack Example',
                filename: entry + '.html',
                template: './src/templates/' + entry + '.ejs',
                chunks: ['commons', 'main']
            }),

            new ExtractTextPlugin({
                filename: entry + '-style.min.css',
                disable: process.env.NODE_ENV === 'development'
            })
        ]
    };
});

module.exports = config;
