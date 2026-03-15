import BudgetSelector from './BudgetSelector'

function Field({ label, id, required, children }) {
  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={id} className="text-sm font-medium text-gray-700">
        {label}{required && <span className="text-purple-500 ml-0.5">*</span>}
      </label>
      {children}
    </div>
  )
}

const inputCls = 'w-full rounded-lg border border-gray-200 px-3 py-2 text-sm text-gray-800 focus:outline-none focus:ring-2 focus:ring-purple-400'

export default function RecipientForm({ values, onChange }) {
  function set(key) {
    return e => onChange({ ...values, [key]: e.target.value })
  }

  return (
    <div className="flex flex-col gap-4">
      <Field label="Recipient's name" id="name" required>
        <input id="name" className={inputCls} value={values.name} onChange={set('name')} placeholder="Alex" required />
      </Field>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <Field label="Your relationship" id="relation">
          <input id="relation" className={inputCls} value={values.relation} onChange={set('relation')} placeholder="best friend, sister…" />
        </Field>
        <Field label="Gender (optional)" id="gender">
          <select id="gender" className={inputCls} value={values.gender} onChange={set('gender')}>
            <option value="">Prefer not to say</option>
            <option value="male">Male</option>
            <option value="female">Female</option>
            <option value="non-binary">Non-binary</option>
          </select>
        </Field>
      </div>

      <Field label="Occasion" id="occasion" required>
        <input id="occasion" className={inputCls} value={values.occasion} onChange={set('occasion')} placeholder="birthday, anniversary, Christmas…" required />
      </Field>

      <Field label="Budget" id="budget">
        <BudgetSelector value={values.budgetTier} onChange={tier => onChange({ ...values, budgetTier: tier })} />
      </Field>
    </div>
  )
}
