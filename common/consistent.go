// 分布式一致性hash算法
package common

import (
	"errors"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

// 自定义切片类型，实现一些基本方法(主要是实现sort接口，实现排序)
type uints []uint32

// 返回长度
func (u uints) Len() int {
	return len(u)
}

// 比较切片中两个数大小(通过索引位置)
func (u uints) Less(i, j int) bool {
	return u[i] < u[j]
}

// 交换切片中的两个数的位置（通过索引位置）
func (u uints) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}

// 定义一个同一的错误信息
var errEmpty = errors.New("Hash 环上没有数据")

// 定义一个保存一致性hash信息的结构体
type Consistent struct {
	// hash集合
	circle map[uint32]string
	// 已经排序的节点hash切片，利用切片有序特性，将hash集合逻辑化为环状结构
	sortedHashes uints
	// 虚拟节点个数
	VituralNode int
	// map读写锁
	sync.RWMutex
}

func NewConsistent() *Consistent {
	return &Consistent{
		circle:      make(map[uint32]string),
		VituralNode: 20,
	}
}

// 默认生产key的方法
func (c *Consistent) generateKey(element string, index int) string {
	return element + strconv.Itoa(index)
}

// 根据key获取hash位置
func (c *Consistent) hashKey(key string) uint32 {
	// // 凑一个64位数组
	// if len(key) < 64 {
	// 	var srcatch [64]byte
	// 	copy(srcatch[:], key)
	// 	// 使用crc32计算校验和
	// 	return crc32.ChecksumIEEE(srcatch[:len(key)])
	// }
	return crc32.ChecksumIEEE([]byte(key))
}

// 更新排序
func (c *Consistent) updateSortedHashes() {
	//获取排序字段，并复位(利用切片的指针特性，直接操作结构体对应字段的内存)
	hashes := c.sortedHashes[:0]
	// 判断hashes是否过大，过大则重置
	if cap(hashes)/(c.VituralNode*4) > len(c.circle) {
		hashes = nil //抛弃原内存
	}
	//将circle集合中节点key，加入hashed中
	for key := range c.circle {
		hashes = append(hashes, key)
	}
	// 排序，方便之后二分查找
	sort.Sort(hashes)
	// 赋值回对应字段
	c.sortedHashes = hashes
}

//添加节点
func (c *Consistent) add(element string) {
	//循环虚拟节点，设置副本
	for i := 0; i < c.VituralNode; i++ {
		//根据生成的节点添加到hash环中
		c.circle[c.hashKey(c.generateKey(element, i))] = element
	}
	//调用排序
	c.updateSortedHashes()
}

// 像hash环中添加节点，注意map线程不安全
func (c *Consistent) Add(element string) {
	c.Lock()
	defer c.Unlock()
	c.add(element)
}

// 删除节点
func (c *Consistent) remove(element string) {
	for i := 0; i < c.VituralNode; i++ {
		delete(c.circle, c.hashKey(c.generateKey(element, i)))
	}
	// 删完要更新排序
	c.updateSortedHashes()
}

// 线程安全的删除节点
func (c *Consistent) Remove(element string) {
	c.Lock()
	defer c.Unlock()
	c.remove(element)
}

//顺时针查找最近的节点
func (c *Consistent) search(key uint32) int {
	f := func(i int) bool {
		return c.sortedHashes[i] >= key
	}
	// 二分查找整个环范围内，最小的index,满足f为true
	return sort.Search(len(c.sortedHashes), f)
}

// 线程安全的获取服务器节点信息(根据请求自带的字符串查找，肯定不能完全匹配，需使用二分查找找到最近的点)
func (c *Consistent) Get(name string) (string, error) {
	c.Lock()
	defer c.Unlock()
	if len(c.circle) == 0 {
		return "", errEmpty
	}
	index := c.search(c.hashKey(name))
	return c.circle[c.sortedHashes[index]], nil
}
