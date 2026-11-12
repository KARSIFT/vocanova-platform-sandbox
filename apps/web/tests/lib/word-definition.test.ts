import assert from "node:assert/strict";
import test from "node:test";
import { getAdditionalDefinition, getPrimaryDefinition } from "../../src/lib/word-definition";

test("plain explanations change display text without mutating canonical identity", () => {
 const meaning = { shortDefinition: "forsake, leave behind", learnerDefinition: "To leave someone or something behind." };
 assert.equal(getPrimaryDefinition(meaning.shortDefinition, meaning.learnerDefinition), "To leave someone or something behind.");
 assert.equal(getPrimaryDefinition("A clear meaning.", " A   CLEAR meaning. "), "A clear meaning.");
 assert.equal(meaning.shortDefinition, "forsake, leave behind");
 assert.equal(getPrimaryDefinition(meaning.shortDefinition, " \n "), meaning.shortDefinition);
 assert.equal(getPrimaryDefinition(meaning.shortDefinition), meaning.shortDefinition);
 assert.equal(getAdditionalDefinition("A clear meaning.", " a  CLEAR meaning. "), null);
});
