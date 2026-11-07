// Newsletter website templates. Each has its own home page layout and
// stylesheet; issue and confirmation pages share markup every template dresses.

import { base, type Template } from "./base";
import { editorial } from "./editorial";
import { almanac } from "./almanac";
import { bento } from "./bento";
import { cinema } from "./cinema";
import { creator } from "./creator";
import { gallery } from "./gallery";
import { soft } from "./soft";
import { terminal } from "./terminal";
import { zine } from "./zine";

export * from "./base";

export const templates: Record<string, Template> = {
  ...editorial,
  bento,
  gallery,
  cinema,
  creator,
  soft,
  zine,
  terminal,
  almanac,
};

export const template = (name: string) => templates[name] || templates.paper;

export const stylesheet = (name: string) => base + template(name).css;

export const grainy = (name: string) => !!template(name).grain;
