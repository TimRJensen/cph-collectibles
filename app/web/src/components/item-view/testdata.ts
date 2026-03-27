export const generic = "data:image/svg+xml," +
  encodeURIComponent(`
    <svg xmlns="http://www.w3.org/2000/svg" width="400" height="250" viewBox="0 0 400 250">
      <rect width="100%" height="100%" fill="#e5e7eb"/>
      <text x="50%" y="50%" dominant-baseline="middle" text-anchor="middle"
            font-family="system-ui,sans-serif" font-size="20" fill="#6b7280">
        Placeholder
      </text>
    </svg>
  `);

export const important = "data:image/svg+xml," +
  encodeURIComponent(`
    <svg xmlns="http://www.w3.org/2000/svg" width="400" height="250" viewBox="0 0 400 250">
      <rect width="100%" height="100%" fill="#e5e7eb"/>
      <text x="50%" y="50%" dominant-baseline="middle" text-anchor="middle"
            font-family="system-ui,sans-serif" font-size="20" fill="#6b7280">
        Highlighted
      </text>
    </svg>
  `);

export const gallery = [
  "data:image/svg+xml," + encodeURIComponent(`
    <svg xmlns="http://www.w3.org/2000/svg" width="400" height="250" viewBox="0 0 400 250">
      <rect width="100%" height="100%" fill="#e5e7eb"/>
      <text x="50%" y="50%" dominant-baseline="middle" text-anchor="middle"
            font-family="system-ui,sans-serif" font-size="20" fill="#6b7280">
        Placeholder
      </text>
    </svg>
  `),
  "data:image/svg+xml," + encodeURIComponent(`
    <svg xmlns="http://www.w3.org/2000/svg" width="400" height="250" viewBox="0 0 400 250">
      <rect width="100%" height="100%" fill="#e5e7eb"/>
      <text x="50%" y="50%" dominant-baseline="middle" text-anchor="middle"
            font-family="system-ui,sans-serif" font-size="20" fill="#6b7280">
        Condition A
      </text>
    </svg>
  `),
  "data:image/svg+xml," + encodeURIComponent(`
    <svg xmlns="http://www.w3.org/2000/svg" width="400" height="250" viewBox="0 0 400 250">
      <rect width="100%" height="100%" fill="#e5e7eb"/>
      <text x="50%" y="50%" dominant-baseline="middle" text-anchor="middle"
            font-family="system-ui,sans-serif" font-size="20" fill="#6b7280">
        Condition B
      </text>
    </svg>
  `),
  "data:image/svg+xml," + encodeURIComponent(`
    <svg xmlns="http://www.w3.org/2000/svg" width="400" height="250" viewBox="0 0 400 250">
      <rect width="100%" height="100%" fill="#e5e7eb"/>
      <text x="50%" y="50%" dominant-baseline="middle" text-anchor="middle"
            font-family="system-ui,sans-serif" font-size="20" fill="#6b7280">
        Condition C
      </text>
    </svg>
  `),
  "data:image/svg+xml," + encodeURIComponent(`
    <svg xmlns="http://www.w3.org/2000/svg" width="400" height="250" viewBox="0 0 400 250">
      <rect width="100%" height="100%" fill="#e5e7eb"/>
      <text x="50%" y="50%" dominant-baseline="middle" text-anchor="middle"
            font-family="system-ui,sans-serif" font-size="20" fill="#6b7280">
        Condition D
      </text>
    </svg>
  `),
];

export const data = [
    {
        id: "a",
        name: "foo",
        url: "",
        caption: "foobar"
    },
    {
        id: "b",
        name: "foo",
        url: "",
        caption: "foobar",
        important: true,

    },
    {
        id: "c",
        name: "foo",
        url: "",
        caption: "foobar"
    },
    {
        id: "d",
        name: "foo",
        url: "",
        caption: "foobar",
    },
    {
        id: "e",
        name: "foo",
        url: "",
        caption: "foobar"
    },
    {
        id: "f",
        name: "foo",
        url: "",
        caption: "foobar"
    },
    {
        id: "g",
        name: "foo",
        url: "",
        caption: "foobar"
    },
    {
        id: "h",
        name: "foo",
        url: "",
        caption: "foobar"
    },
    {
        id: "i",
        name: "foo",
        url: "",
        caption: "foobar",
        important: true,

    },
    {
        id: "h",
        name: "foo",
        url: "",
        caption: "foobar"
    },
    {
        id: "h",
        name: "foo",
        url: "",
        caption: "foobar"
    },
    {
        id: "h",
        name: "foo",
        url: "",
        caption: "foobar"
    },
];

export default data;