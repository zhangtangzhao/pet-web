export interface CategoryItem {
  id: string
  name: string
  icon: string
  breedCount: number
}

export interface ProductCard {
  id: string
  title: string
  mainImage: string
  price: string
  originalPrice: string
  breedName: string
  petGender: number
  favoriteCount: number
  sales: number
  status: number
}

export interface HomeResp {
  categories: CategoryItem[]
  hot: ProductCard[]
}

export interface PetProfile {
  gender: number
  genderText: string
  birthDate: string
  ageText: string
  vaccineDesc: string
  dewormDesc: string
  bodyType: string
  coatColor: string
  personality: string
  healthDesc: string
}

export interface ProductDetail extends ProductCard {
  images: string[]
  videoUrl: string
  videoCover: string
  detailHtml: string
  petProfile: PetProfile
  breed: { id: string; categoryId: string; name: string; cover: string }
  category: { id: string; name: string }
  isFavorite: boolean
  aiEnabled?: boolean
}

export interface AIChatMessage {
  role: 'user' | 'assistant'
  content: string
}

export interface AIKnowledgeRef {
  title: string
  content: string
  isBreedSpecific: boolean
}

export interface AIAskResp {
  answer: string
  refs: AIKnowledgeRef[]
  demo: boolean
}

export interface PageResp<T> {
  total: number
  list: T[]
}

export interface LoginResp {
  accessToken: string
  refreshToken: string
  expiresIn: number
  member: { id: string; nickname: string; avatar: string; phone: string; isNew: boolean }
}

export interface OrderView {
  orderNo: string
  status: number
  statusText: string
  payAmount: string
  contactName: string
  contactPhone: string
  expireAt: string
  createdAt: string
  items: { productId: string; productTitle: string; productImage: string; breedName: string; price: string; quantity: number }[]
}
