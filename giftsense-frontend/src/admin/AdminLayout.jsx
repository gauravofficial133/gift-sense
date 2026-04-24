import { Link, Outlet, useLocation } from 'react-router-dom'
import { LayoutTemplate, Image, Gift } from 'lucide-react'

const navItems = [
  { to: '/admin/templates', label: 'Templates', icon: LayoutTemplate },
  { to: '/admin/assets', label: 'Assets', icon: Image },
]

export default function AdminLayout() {
  const location = useLocation()

  return (
    <div className="min-h-screen bg-gray-50 flex">
      <aside className="w-56 bg-white border-r border-gray-200 flex flex-col">
        <Link to="/" className="flex items-center gap-2 px-4 py-4 border-b border-gray-100">
          <Gift className="w-5 h-5 text-orange-500" />
          <span className="text-sm font-bold text-gray-900">upahaar.ai</span>
          <span className="text-[10px] text-gray-400 ml-auto">admin</span>
        </Link>
        <nav className="flex-1 py-3">
          {navItems.map(({ to, label, icon: Icon }) => {
            const active = location.pathname.startsWith(to)
            return (
              <Link
                key={to}
                to={to}
                className={`flex items-center gap-2 px-4 py-2 text-sm transition-colors ${
                  active ? 'text-orange-600 bg-orange-50 font-medium' : 'text-gray-600 hover:bg-gray-50'
                }`}
              >
                <Icon className="w-4 h-4" />
                {label}
              </Link>
            )
          })}
        </nav>
      </aside>
      <main className="flex-1 overflow-auto">
        <Outlet />
      </main>
    </div>
  )
}
