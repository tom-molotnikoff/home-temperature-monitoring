export function linesHiddenReducer(
  state: { [key: string]: boolean },
  action: { type: "toggle" | "reset"; key: string }
) {
  switch (action.type) {
    case "toggle":
      return { ...state, [action.key]: !state[action.key] };
    case "reset":
      return { ...state, [action.key]: false };
  }
}
