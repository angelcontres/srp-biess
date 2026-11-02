# Campo de Formulario

Estructura y directrices para inputs, selects, textareas y controles de formulario accesibles.

---

## Estructura de Campo

```tsx
<div>
  <label htmlFor="cedula" className="block text-base font-medium text-neutral-900">
    Cédula de identidad <span className="font-normal text-muted">(obligatorio)</span>
  </label>
  <p id="cedula-ayuda" className="mt-1 text-sm text-muted">Ingrese los 10 números de su cédula.</p>
  <input
    id="cedula"
    inputMode="numeric"
    autoComplete="username"
    aria-describedby="cedula-ayuda cedula-error"
    aria-invalid={hasError}
    className="mt-2 block w-full h-11 px-3 rounded-md border border-line-input bg-surface text-base text-neutral-900
               focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark
               aria-[invalid=true]:border-2 aria-[invalid=true]:border-danger"
  />
  {hasError && (
    <p id="cedula-error" className="mt-2 flex gap-2 text-base text-danger">
      <CircleAlert size={20} aria-hidden="true" className="shrink-0" /> La cédula debe tener 10 números.
    </p>
  )}
</div>
```

---

## Tratamiento por Tipo de Campo

| Tipo | Tratamiento |
|---|---|
| **Obligatorio** | Texto explícito "(obligatorio)" dentro de la etiqueta `<label>` + atributo nativo `required`. Nunca usar únicamente un asterisco (`*`). |
| **Verificado** | `bg-neutral-100`, sin borde interactivo, `readOnly`, badge "Verificado" (`ShieldCheck`) y texto de ayuda: *"Dato verificado. No se puede modificar."* |
| **Solo lectura** | `bg-neutral-100`, sin borde interactivo, `readOnly`, sin badge. |
| **Con error** | Borde de 2 px `danger` + mensaje descriptivo con icono `CircleAlert` ubicado debajo del input. |
| **Calculado** | Valor mostrado en texto + caption *"Calculado automáticamente"*, envuelto en contenedor con `aria-live="polite"`. |

---

## Reglas de Implementación

- **Placeholders:** Nunca usar el atributo `placeholder` como sustituto de la etiqueta `<label>`.
- **Momento de validación:** Validar al perder el foco (`onBlur`) o al guardar, **nunca mientras el usuario escribe**.
- **Campos de fecha:** Utilizar tres campos separados (día, mes, año) o un input de texto con formato visual `dd/mm/aaaa`. No depender del selector nativo del navegador.
- **Cédula ecuatoriana:** Atributo `inputMode="numeric"`, nunca `type="number"`.
- **Portapapeles:** Se debe permitir la acción de pegar en todos los campos sin restricciones (incluyendo contraseñas).
- **Selecciones múltiples y booleanas:** Grupos de radio o checkbox siempre contenidos dentro de `<fieldset>` y `<legend>`. Sin opciones preseleccionadas por defecto en declaraciones o consentimientos. **Prohibido utilizar interruptores (toggles)** para preguntas de Sí/No.
