// API and service tests run in Node; DOM matchers are not needed here.
// jsdom has no TextEncoder; email size checks need it.
globalThis.TextEncoder ??= require("util").TextEncoder;
