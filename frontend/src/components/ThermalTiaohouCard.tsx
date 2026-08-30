import { Thermometer } from 'lucide-react'
import './ThermalTiaohouCard.css'

interface ThermalTiaohouCardProps {
  thermal: {
    status: string
    condition: string
    required_elements?: string
    visible_support: string[]
    hidden_support: string[]
    detail: string
  }
}

const statusLabel: Record<string, string> = {
  urgent_resolved: '急需已透',
  urgent_partial: '急需有根待引',
  urgent_unresolved: '急需未解',
  seasonal_resolved: '季候已有调剂',
  seasonal_partial: '季候藏支可用',
  seasonal_unresolved: '季候调剂不足',
  non_urgent: '寒热无急',
}

const ThermalTiaohouCard = ({ thermal }: ThermalTiaohouCardProps) => {
  const visible = thermal.visible_support || []
  const hidden = thermal.hidden_support || []

  return (
    <div className="thermal-tiaohou-card">
      <div className="thermal-tiaohou-card__header">
        <span><Thermometer size={15} />寒热调候</span>
        <strong>{statusLabel[thermal.status] || '寒热待查'}</strong>
      </div>
      <div className="thermal-tiaohou-card__facts">
        <span>季候：{thermal.condition || '平和'}</span>
        <span>调剂：{thermal.required_elements || '无急需'}</span>
        <span>透：{visible.length ? visible.join('、') : '无'}</span>
        <span>藏：{hidden.length ? hidden.join('、') : '无'}</span>
      </div>
      <p>{thermal.detail}</p>
    </div>
  )
}

export default ThermalTiaohouCard
