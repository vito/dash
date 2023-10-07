/**
 * @file Dash grammar for tree-sitter
 * @author Alex Suraci <suraci.alex@gmail.com>
 * @author Amaan Qureshi <amaanq12@gmail.com>
 * @license MIT
 * @see {@link https://bass-lang.org|official website}
 */

/* eslint-disable arrow-parens */
/* eslint-disable camelcase */
/* eslint-disable-next-line spaced-comment */
/// <reference types="tree-sitter-cli/dsl" />
// @ts-check

var fs = require('fs');

module.exports = JSON.parse(
  fs.readFileSync('src/grammar.json', 'utf8')
);
