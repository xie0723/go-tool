package stringmatch

// 定义堆栈的值范围，泛型
type StackValue interface {
	~string | ~bool
}

// 栈 对象，用来存储数据或者比较结果
type Stack[T StackValue] struct {
	size int //栈最大可以存放的数量
	top  int //栈顶
	data []T //模拟栈
}

// 实例话一个新堆栈
func newStack[T StackValue](t T, size int) *Stack[T] {
	return &Stack[T]{
		size: size,
		top:  0,
		data: make([]T, size),
	}
}

// 判断堆栈是否满了
func (s *Stack[T]) IsFull() bool {
	return s.top == s.size
}

// 判断堆栈是否为空
func (s *Stack[T]) IsEmpty() bool {
	return s.top == 0
}

// 取出栈值-后进先出
func (s *Stack[T]) Pop() T {
	if s.IsEmpty() {
		panic("pop err, stack is empty")
	}
	s.top--
	return s.data[s.top]
}

// 压入栈值
func (s *Stack[T]) Push(d T) {
	if s.IsFull() {
		panic("push err, stack is full")
	}
	s.data[s.top] = d
	s.top++
}

// 获得栈值，但是不出栈
func (s *Stack[T]) Peek() T {
	if s.IsEmpty() {
		panic("peek err, stack is empty")
	}
	return s.data[s.top-1]
}

// 计算栈内还是多少个值
func (s *Stack[T]) Len() int {
	return s.top
}
