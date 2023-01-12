module.exports = {
  content: ["./templates/**/*.html"],
  theme: {
    extend: {
      fontFamily: {
        "sans": ["Inter", "system-ui", "sans-serif"],
        "mono": ["Source Code Pro", "monospace"]
      },
      colors: {
        "brand": {
          500: "#0069ff",
          900: "#031b4d"
        }
      }
    }
  },
  plugins: []
}
