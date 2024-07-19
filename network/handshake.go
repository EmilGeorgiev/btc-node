package network

//func Connect(addr string) {
//	conn, err := net.Dial("tcp", addr)
//	if err != nil {
//		fmt.Println("Error:", err)
//		return
//	}
//	defer conn.Close()
//	fmt.Printf("opened a connection with remote address: %s\n", conn.RemoteAddr().String())
//
//	if err = initializeHandshake(conn); err != nil {
//		return
//	}
//}

//func initializeHandshake(conn net.Conn) error {
//	if err := sendVersionMsg(conn); err != nil {
//		return err
//	}
//
//	if err := readVersionMsg(conn); err != nil {
//		return err
//	}
//
//	if err := readVerackMsg(conn); err != nil {
//		return err
//	}
//
//	if err := sendVerackMsg(conn); err != nil {
//		return err
//	}
//}
//
//func sendVersionMsg(conn net.Conn) error {
//	conn.Write()
//}
