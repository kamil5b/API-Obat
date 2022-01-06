package routes

import (
	"github.com/gofiber/fiber"
	"github.com/kamil5b/API-Obat/controllers"
)

func Setup(app *fiber.App) {
	//192.168.1.18:8000
	app.Get("/", func(c *fiber.Ctx) error {

		return c.JSON(fiber.Map{
			"message": "success",
		})
	})
	//----AUTH----
	app.Post("/api/register", controllers.Register)
	app.Post("/api/login", controllers.Login)
	app.Get("/api/user", controllers.GetSalesman)
	app.Post("/api/user", controllers.PostUser)
	app.Post("/api/logout", controllers.Logout)

	//----BANK----
	app.Get("/api/bank", controllers.GetBank)

	//----CUSTOMER----
	app.Get("/api/customer", controllers.GetCustomer)
	app.Post("/api/customer", controllers.PostCustomer)
	app.Put("/api/customer", controllers.UpdateCustomer)

	//----TOKO----
	app.Get("/api/toko", controllers.GetToko)
	app.Post("/api/toko", controllers.PostToko)
	app.Put("/api/toko", controllers.UpdateToko)

	//----GIRO----
	app.Get("/api/giro", controllers.GetGiro)
	app.Post("/api/giro", controllers.PostGiro)

	//----STOCK----
	app.Get("/api/stock/summary", controllers.GetStockSummary)
	app.Get("/api/stock", controllers.GetAllStock)
	app.Put("/api/stock/unbox", controllers.UnboxStock)

	//----RETUR----
	app.Get("/api/retur", controllers.GetAllRetur)
	app.Post("/api/retur", controllers.ReturBarang)

	//----PIUTANG----
	app.Get("/api/piutang", controllers.GetPiutang)
	app.Get("/api/piutang/record", controllers.GetRecordPiutang)
	app.Put("/api/piutang/bayar", controllers.BayarPiutang)
	app.Put("/api/piutang/naik", controllers.PiutangNaik)

	//----HUTANG----
	app.Get("/api/hutang/record", controllers.GetRecordHutang)
	app.Get("/api/hutang", controllers.GetHutang)
	app.Put("/api/hutang/bayar", controllers.BayarHutang)
	app.Put("/api/hutang/naik", controllers.HutangNaik)

	//----PENJUALAN----
	app.Get("/api/penjualan", controllers.GetAllPenjualan)
	app.Post("/api/penjualan", controllers.JualBarang)
	app.Post("/api/penjualan/barang", controllers.PenjualanPerBarang)
	app.Post("/api/penjualan/faktur", controllers.GetPenjualanPerFaktur)
	//----SUMMARY PENJUALAN----
	app.Get("/api/penjualan/summary", controllers.SummaryPenjualan)
	app.Post("/api/penjualan/summary/faktur", controllers.SummaryPenjualanPerFaktur)
	app.Post("/api/penjualan/summary/tanggal", controllers.SummaryPenjualanPerTanggal)
	app.Get("/api/penjualan/summary/barang", controllers.SummaryPenjualanPerBarang)
	app.Post("/api/penjualan/summary/barang", controllers.SummaryPenjualanPerBarangTanggal)
	app.Post("/api/penjualan/summary", controllers.TotalSummaryPenjualan)

	//----PEMBELIAN----
	app.Get("/api/pembelian", controllers.GetAllPembelian)
	app.Post("/api/pembelian", controllers.BeliBarang)
	app.Post("/api/pembelian/barang", controllers.PembelianPerBarang)

	//----SUMMARY PEMBELIAN----
	app.Get("/api/pembelian/summary", controllers.SummaryPembelian)
	app.Post("/api/pembelian/summary", controllers.TotalSummaryPembelian)
	app.Post("/api/pembelian/summary/faktur", controllers.SummaryPembelianPerFaktur)
	app.Post("/api/pembelian/summary/tanggal", controllers.SummaryPembelianPerTanggal)
	app.Get("/api/pembelian/summary/barang", controllers.SummaryPembelianPerBarang)

	//----FAKTUR----
	app.Post("/api/faktur/pembelian", controllers.FakturPembelian)
	app.Get("/api/faktur/pembelian", controllers.GetFakturPembelian)
	app.Post("/api/faktur/penjualan", controllers.FakturPenjualan)
	app.Get("/api/faktur/penjualan", controllers.GetFakturPenjualan)

	//----BARANG----
	app.Post("/api/barang", controllers.BarangBaru)
	app.Post("/api/barang/kode", controllers.GetSatuBarang)
	app.Get("/api/barang", controllers.GetAllBarang)
	app.Post("/api/barang/update", controllers.PengajuanUpdateBarang)

	//----SUPERVISOR----
	app.Get("/api/target", controllers.GetTargetSales)
	app.Post("/api/target", controllers.PostTargetSales)
	app.Put("/api/target", controllers.GantiStatus)

	app.Get("/api/barang/update", controllers.GetPengajuan)
	app.Put("/api/barang/update", controllers.UpdateBarang)
	app.Delete("/api/barang/update", controllers.TolakPengajuan)

	//----SALES-----
	app.Get("/api/sales", controllers.GetAllSales)
	app.Post("/api/sales", controllers.GetIndividualSales)
	app.Put("/api/sales", controllers.EditInsentif)

	//----DEMANDS----
	app.Get("/api/demand", controllers.GetAllDemand)
	app.Delete("/api/demand", controllers.DeleteDemand)

	//----PROFIT----
	app.Get("/api/profit", controllers.GetProfits)

}
