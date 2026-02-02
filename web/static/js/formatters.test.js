import { expect, test } from "vitest";
import { formatCPF } from "./formatters.js";

test("formatCPF remove caracteres não numéricos", () => {
  expect(formatCPF("abc")).toBe("");
});

test("formatCPF formata corretamente dígitos", () => {
  expect(formatCPF("12345678901")).toBe("123.456.789-01");
});

test("formatCPF formata progressivamente", () => {
  expect(formatCPF("1")).toBe("1");
  expect(formatCPF("123")).toBe("123");
  expect(formatCPF("1234")).toBe("123.4");
  expect(formatCPF("123456")).toBe("123.456");
  expect(formatCPF("123456789")).toBe("123.456.789");
  expect(formatCPF("12345678901")).toBe("123.456.789-01");
});
