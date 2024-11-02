const palette = require('tailwindcss/colors');

/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ["./web/**/*.{js,html}"],
    theme: {
        fontFamily: {
            sans: ['Inter', 'sans-serif'],
        },
        colors: {
            black: palette.black,
            white: palette.white,
            info: palette.sky,
            error: palette.red,
            success: palette.green,
            warning: palette.orange,
            primary: palette.blue,
            secondary: palette.slate,
            transparent: palette.transparent,
        },
    },
    plugins: [
        require('@tailwindcss/typography'),
        require('@tailwindcss/forms'),
    ],
};
