// Snowpack Configuration File
// See all supported options: https://www.snowpack.dev/reference/configuration

/** @type {import("snowpack").SnowpackUserConfig } */
module.exports = {
  mount: {
    src: '/',
    public: { url: '/', static: true, resolve: false },
  },
  plugins: [
    '@snowpack/plugin-sass',
    '@snowpack/plugin-vue',
  ],
  packageOptions: {
    /* ... */
  },
  devOptions: {
    open: "none"
  },
  buildOptions: {
    out: '../app/public'
  },
  routes: [
    {
      src: '/(api|server)/.*',
      dest: (req, res) => {
        return require('http2-proxy').web(req, res, {
          hostname: 'localhost',
          port: 25459,
        });
      },
    },
  ]
};
