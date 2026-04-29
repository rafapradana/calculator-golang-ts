const expressionEl = document.getElementById("expression") as HTMLElement;
const resultEl = document.getElementById("result") as HTMLElement;

const OPERATORS = ["+", "-", "*", "/"];

let expression = "";
let lastResult: string | null = null;

function updateDisplay(): void {
  expressionEl.textContent = expression;
}

function appendValue(value: string): void {
  const lastChar = expression.slice(-1);

  if (lastResult !== null && !OPERATORS.includes(value)) {
    expression = "";
    lastResult = null;
  }

  if (OPERATORS.includes(value) && expression === "") {
    if (value === "-") {
      expression = value;
    }
    updateDisplay();
    return;
  }

  if (OPERATORS.includes(value) && OPERATORS.includes(lastChar)) {
    expression = expression.slice(0, -1) + value;
    updateDisplay();
    return;
  }

  if (value === "." && expression !== "") {
    const parts = expression.split(/[\+\-\*\/]/);
    const currentNumber = parts[parts.length - 1];
    if (currentNumber.includes(".")) {
      return;
    }
  }

  expression += value;
  updateDisplay();
}

function clearAll(): void {
  expression = "";
  lastResult = null;
  resultEl.textContent = "0";
  updateDisplay();
}

function backspace(): void {
  if (lastResult !== null) {
    clearAll();
    return;
  }
  expression = expression.slice(0, -1);
  if (expression === "") {
    resultEl.textContent = "0";
  }
  updateDisplay();
}

async function calculate(): Promise<void> {
  if (expression === "") return;

  try {
    const response = await fetch("/api/calculate", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ expression }),
    });

    const data = await response.json();

    if (data.error) {
      resultEl.textContent = "Error";
      lastResult = null;
      return;
    }

    const formatted = parseFloat(data.result.toFixed(10)).toString();

    resultEl.textContent = formatted;
    lastResult = formatted;
  } catch {
    resultEl.textContent = "Error";
    lastResult = null;
  }
}

document.querySelectorAll("button").forEach((button) => {
  button.addEventListener("click", () => {
    if (button.dataset.action === "clear") {
      clearAll();
      return;
    }

    if (button.dataset.action === "backspace") {
      backspace();
      return;
    }

    if (button.dataset.action === "calculate") {
      calculate();
      return;
    }

    if (button.dataset.value) {
      appendValue(button.dataset.value);
    }
  });
});

document.addEventListener("keydown", (e: KeyboardEvent) => {
  if (/^[0-9.]$/.test(e.key)) {
    appendValue(e.key);
  } else if (["+", "-", "*", "/"].includes(e.key)) {
    appendValue(e.key);
  } else if (e.key === "Enter" || e.key === "=") {
    e.preventDefault();
    calculate();
  } else if (e.key === "Backspace") {
    backspace();
  } else if (e.key === "Escape") {
    clearAll();
  }
});
